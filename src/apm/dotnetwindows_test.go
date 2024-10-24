package apm

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/andrew-lozoya/k8s-agents-operator-windows/src/api/v1alpha2"
)

func TestDotnetWindowsInjector_Language(t *testing.T) {
	require.Equal(t, "dotnet-windows", (&DotnetWindowsInjector{}).Language())
}

func TestDotnetWindowsInjector_Inject(t *testing.T) {
	vtrue := true
	tests := []struct {
		name           string
		pod            corev1.Pod
		ns             corev1.Namespace
		inst           v1alpha2.Instrumentation
		expectedPod    corev1.Pod
		expectedErrStr string
	}{
		{
			name: "nothing",
		},
		{
			name: "a container, no instrumentation",
			pod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			expectedPod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
		},
		{
			name: "a container, wrong instrumentation (not the correct lang)",
			pod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			expectedPod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			inst: v1alpha2.Instrumentation{Spec: v1alpha2.InstrumentationSpec{Agent: v1alpha2.Agent{Language: "not-this"}}},
		},
		{
			name: "a container, instrumentation with blank licenseKeySecret",
			pod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			expectedPod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			expectedErrStr: "licenseKeySecret must not be blank",
			inst:           v1alpha2.Instrumentation{Spec: v1alpha2.InstrumentationSpec{Agent: v1alpha2.Agent{Language: "dotnet"}}},
		},
		{
			name: "a container, instrumentation",
			pod: corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "test"},
			}}},
			expectedPod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "test",
					Env: []corev1.EnvVar{
						{Name: "CORECLR_ENABLE_PROFILING", Value: "1"},
						{Name: "CORECLR_PROFILER", Value: "{36032161-FFC0-4B61-B559-F6C5D41BAE5A}"},
						{Name: "CORECLR_PROFILER_PATH", Value: "C:\\newrelic-instrumentation\\netcore\\NewRelic.Profiler.dll"},
						{Name: "CORECLR_NEWRELIC_HOME", Value: "C:\\newrelic-instrumentation\\netcore"},
						{Name: "COR_ENABLE_PROFILING", Value: "1"},
						{Name: "COR_PROFILER", Value: "{71DA0A04-7777-4EC6-9643-7D28B46A8A41}"},
						{Name: "COR_PROFILER_PATH", Value: "C:\\newrelic-instrumentation\\netframework\\NewRelic.Profiler.dll"},
						{Name: "NEWRELIC_HOME", Value: "C:\\newrelic-instrumentation\\netframework"},
						{Name: "NEW_RELIC_APP_NAME", Value: "test"},
						{Name: "NEW_RELIC_LICENSE_KEY", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "newrelic-key-secret"}, Key: "new_relic_license_key", Optional: &vtrue}}},
						{Name: "NEW_RELIC_LABELS", Value: "operator:auto-injection"},
						{Name: "NEW_RELIC_K8S_OPERATOR_ENABLED", Value: "true"},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: "newrelic-instrumentation", MountPath: "C:\\newrelic-instrumentation"}},
				}},
				InitContainers: []corev1.Container{{
					Name:         "newrelic-instrumentation-dotnet-windows",
					Command:      []string{"xcopy", "C:\\instrumentation", "C:\\newrelic-instrumentation", "/E", "/I", "/H", "/Y"},
					VolumeMounts: []corev1.VolumeMount{{Name: "newrelic-instrumentation", MountPath: "C:\\newrelic-instrumentation"}},
				}},
				Volumes: []corev1.Volume{{Name: "newrelic-instrumentation", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			}},
			inst: v1alpha2.Instrumentation{Spec: v1alpha2.InstrumentationSpec{Agent: v1alpha2.Agent{Language: "dotnet-windows"}, LicenseKeySecret: "newrelic-key-secret"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			i := &DotnetWindowsInjector{}
			actualPod, err := i.Inject(ctx, test.inst, test.ns, test.pod)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			require.Equal(t, test.expectedErrStr, errStr)
			if diff := cmp.Diff(test.expectedPod, actualPod); diff != "" {
				assert.Fail(t, diff)
			}
		})
	}
}
