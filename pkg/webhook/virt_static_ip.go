package webhook

import (
	"context"
	"net/http"
	"strings"

	"github.com/alauda/kube-ovn/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	virtv1 "kubevirt.io/client-go/api/v1"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	vmGVK = metav1.GroupVersionKind{Group: virtv1.GroupName, Version: virtv1.GroupVersion.Version, Kind: "VirtualMachine"}
)

func (v *ValidatingHook) VirtualMachineCreateHook(ctx context.Context, req admission.Request) admission.Response {
	o := virtv1.VirtualMachine{}
	if err := v.decoder.Decode(req, &o); err != nil {
		return ctrlwebhook.Errored(http.StatusBadRequest, err)
	}
	// Get vm template static ips
	staticIPSAnno, ok := o.Spec.Template.ObjectMeta.Annotations[util.IpPoolAnnotation]
	if !ok {

	}
	klog.V(3).Infof("%s %s@%s, ip_pool: %s", o.Kind, o.GetName(), o.GetNamespace(), staticIPSAnno)
	if staticIPSAnno == "" {
		return ctrlwebhook.Allowed("by pass")
	}
	return v.podControllerCreate(ctx, staticIPSAnno, o.GetNamespace())
}

func (v *ValidatingHook) VirtualMachineUpdateHook(ctx context.Context, req admission.Request) admission.Response {
	o := virtv1.VirtualMachine{}
	n := virtv1.VirtualMachine{}
	if err := v.decoder.DecodeRaw(req.OldObject, &o); err != nil {
		return ctrlwebhook.Errored(http.StatusBadRequest, err)
	}
	if err := v.decoder.DecodeRaw(req.Object, &n); err != nil {
		return ctrlwebhook.Errored(http.StatusBadRequest, err)
	}
	// Get pod template static ips
	oldStaticIPSAnno, ok := o.Spec.Template.ObjectMeta.Annotations[util.IpPoolAnnotation]
	newStaticIPSAnno, ok := n.Spec.Template.ObjectMeta.Annotations[util.IpPoolAnnotation]
	if !ok {
	}
	klog.V(3).Infof("%s %s@%s, old ip_pool: %s, new ip_pool:%s", o.Kind, o.GetName(), o.GetNamespace(), oldStaticIPSAnno, newStaticIPSAnno)
	if len(util.DiffStringSlice(strings.Split(oldStaticIPSAnno, ","), strings.Split(newStaticIPSAnno, ","))) == 0 {
		return ctrlwebhook.Allowed("by pass")
	}
	return v.podControllerUpdate(ctx, oldStaticIPSAnno, newStaticIPSAnno, o.GetNamespace())
}

func (v *ValidatingHook) VirtualMachineDeleteHook(ctx context.Context, req admission.Request) admission.Response {
	o := virtv1.VirtualMachine{}
	if err := v.client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name}, &o); err != nil {
		return ctrlwebhook.Errored(http.StatusBadRequest, err)
	}
	staticIPSAnno, ok := o.Spec.Template.ObjectMeta.Annotations[util.IpPoolAnnotation]
	if !ok {
	}
	klog.V(3).Infof("%s %s@%s, ip_pool: %s", o.Kind, o.GetName(), o.GetNamespace(), staticIPSAnno)
	if staticIPSAnno == "" {
		return ctrlwebhook.Allowed("by pass")
	}
	return v.podControllerDelete(ctx, staticIPSAnno, o.GetNamespace())
}
