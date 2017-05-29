/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"net/http"
	"testing"

	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
)

func testv1beta1Data() *extensionsv1beta1.Deployment {
	one := int32(1)
	deployment := &extensionsv1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "foo1",
			Labels:          map[string]string{"app": "foo"},
			Namespace:       "default",
			ResourceVersion: "12345",
		},
		Spec: extensionsv1beta1.DeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "foo"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "app", Image: "abc/app:v4"},
						{Name: "ape", Image: "zyx/ape"}},
				},
			},
		},
	}
	return deployment
}
func TestV1beta1GetObjects(t *testing.T) {
	deployment := testv1beta1Data()

	f, tf, codec, _ := cmdtesting.NewAPIFactory()
	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = &fake.RESTClient{
		APIRegistry:          api.Registry,
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, deployment)},
	}
	tf.Namespace = "default"
	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	cmd := NewCmdGet(f, buf, errBuf)
	cmd.SetOutput(buf)
	cmd.Run(cmd, []string{"deployment", "foo"})

	expected := []runtime.Object{deployment}
	verifyObjects(t, expected, tf.Printer.(*testPrinter).Objects)
	fmt.Println(buf, buf.String())

	if len(buf.String()) == 0 {
		t.Errorf("unexpected empty output")
	}
}

