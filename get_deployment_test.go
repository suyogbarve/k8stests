package cmd

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
