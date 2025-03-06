package iamram

import (
	"context"
	"dancav.io/aws-iamra-manager/api/v1"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"strings"
)

func ReconcilePod(
	ctx context.Context, k *kubernetes.Clientset, kcfg *rest.Config,
	session *v1.AwsIamRaSession, pod corev1.Pod,
) error {
	logger := log.FromContext(ctx)

	command := []string{
		"update-config",
		"-t", string(session.Spec.TrustAnchorArn),
		"-p", string(session.Spec.ProfileArn),
		"-r", string(session.Spec.RoleArn),
		"-d", strconv.Itoa(int(session.Spec.DurationSeconds)),
	}
	roleSessionName := fmt.Sprintf("%s@%s", pod.Namespace, pod.Name)
	if session.Spec.RoleSessionName != "" {
		roleSessionName = session.Spec.RoleSessionName
	}
	command = append(command, "-n", roleSessionName)

	logger.Info("Executing remote command", "command", command)
	execReq := k.CoreV1().RESTClient().
		Post().
		Resource(string(corev1.ResourcePods)).
		Name(pod.Name).
		Namespace(pod.Namespace).
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
		return err
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
		return err
	}
	logger.Info("remote command succeeded!",
		"stdout", stdout.String(), "stderr", stderr.String())

	return nil
}
