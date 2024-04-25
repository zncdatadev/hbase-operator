package v1alpha1

func (in *HbaseCluster) GetLabels() map[string]string {
	labels := map[string]string{
		"app": in.Name,
	}
	return labels
}
