package iamram

import (
	"context"
	"dancav.io/aws-iamra-manager/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"strings"
	"time"
)

const ExpirationBufferSeconds = 60

func ReconcilePod(
	ctx context.Context, k *kubernetes.Clientset, kcfg *rest.Config,
	session *v1.AwsIamRaSession, pod corev1.Pod,
) (*time.Time, bool, error) {
	logger := log.FromContext(ctx)

	podRef := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}
	expiration := getExpirationForPod(session, podRef)
	if !needToRefresh(ctx, expiration) {
		logger.Info("pod already has active session credentials",
			"pod", podRef, "expiration", expiration)
		return expiration, false, nil
	}

	command := []string{
		"update-credentials",
		"-t", string(session.Spec.TrustAnchorArn),
		"-p", string(session.Spec.ProfileArn),
		"-r", string(session.Spec.RoleArn),
	}
	if session.Spec.DurationSeconds > 0 {
		command = append(command, "-d", strconv.Itoa(int(session.Spec.DurationSeconds)))
	}
	roleSessionName := strings.Replace(podRef.String(), "/", "@", 1)
	if session.Spec.RoleSessionName != "" {
		roleSessionName = session.Spec.RoleSessionName
	}
	command = append(command, "-n", roleSessionName)

	logger.Info("Executing remote command", "command", command)
	execReq := k.CoreV1().RESTClient().
		Post().
		Resource(string(corev1.ResourcePods)).
		Name(podRef.Name).
		Namespace(podRef.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "aws-iamra-manager",
			Command:   command,
			Stdout:    true,
			Stderr:    true,
			Stdin:     false,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(kcfg, http.MethodPost, execReq.URL())
	if err != nil {
		logger.Error(err, "unable to create remote executor")
		return nil, false, err
	}

	var stdout strings.Builder
	var stderr strings.Builder

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		logger.Error(err, "remote command execution failed",
			"stdout", stdout.String(), "stderr", stderr.String())
		return nil, false, err
	}
	logger.Info("remote command succeeded!",
		"stdout", stdout.String(), "stderr", stderr.String())

	if session.Status.ExpirationTimes == nil {
		session.Status.ExpirationTimes = make(map[string]metav1.Time)
	}

	newExpiration, err := time.Parse("2006-01-02T15:04:05Z",
		strings.TrimSuffix(stdout.String(), "\n"))
	if err != nil {
		logger.Error(err, "unable to parse IAM session expiration time",
			"expiration", stdout.String())
		return nil, false, err
	}
	session.Status.ExpirationTimes[podRef.String()] = metav1.NewTime(newExpiration)

	return &newExpiration, true, nil
}

func getExpirationForPod(session *v1.AwsIamRaSession, podRef types.NamespacedName) *time.Time {
	if session.Status.ExpirationTimes == nil {
		return nil
	}
	t := session.Status.ExpirationTimes[podRef.String()]
	return &t.Time
}

func needToRefresh(ctx context.Context, expiration *time.Time) bool {
	logger := log.FromContext(ctx)
	if expiration == nil {
		logger.Info("pod does not have active credentials")
		return true
	}
	if time.Now().After(*expiration) {
		logger.Info("credentials are already expired", "expiration", expiration)
		return true
	}
	// The buffer is incorporated once into the requeue interval, and include it a second
	// time by doubling it here, to reduce risk of missing it.
	timeToRefresh := expiration.Add(-2 * ExpirationBufferSeconds * time.Second)
	if time.Now().After(timeToRefresh) {
		logger.Info("refreshing credentials before they expire", "expiration", expiration)
		return true
	}
	return false
}
