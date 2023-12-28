# Velero 1.11.1

- Install
```shell
velero install --provider gcp --plugins velero/velero-plugin-for-gcp:v1.8.2 --bucket mig-controller-demo --secret-file ~/Desktop/forkrepo/mig-controller/deployexample/gcp-credentials --use-volume-snapshots false --use-node-agent
```
- Annotate Restore Pod for Restic
```shell
kubectl -n cluster-migrate-test-app annotate pod/POD_NAME backup.velero.io/backup-volumes=PV_NAME_1,PV_NAME_2
```
- Create Backup
```shell
velero backup create migrate-backup --include-namespaces cluster-migrate-test-app --storage-location default  --volume-snapshot-locations default
```
```shell
velero backup create migrate-backup --include-namespaces cluster-migrate-test-app --storage-location default  --volume-snapshot-locations default --include-resources pods,deployments,pvc,service
```
hoặc
```yaml
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: migrate-backup
  namespace: openshift-migration
spec:
  defaultVolumesToRestic: false
  includedNamespaces:
  - cluster-migrate-test-app
  storageLocation: default
```
```shell
kg backup migrate-backup -n openshift-migration -o yaml
```
```yaml
apiVersion: velero.io/v1
kind: Backup
metadata:
  annotations:
    velero.io/source-cluster-k8s-gitversion: v1.25.10-gke.2700
    velero.io/source-cluster-k8s-major-version: "1"
    velero.io/source-cluster-k8s-minor-version: "25"
  creationTimestamp: "2023-12-13T01:38:46Z"
  generation: 5
  labels:
    velero.io/storage-location: default
  name: migrate-backup
  namespace: openshift-migration
  resourceVersion: "4092523"
  uid: 4580b92f-25ec-482e-900d-242c4a90cec2
spec:
  defaultVolumesToRestic: false
  hooks: {}
  includedNamespaces:
  - cluster-migrate-test-app
  metadata: {}
  storageLocation: default
  ttl: 720h0m0s
  volumeSnapshotLocations:
  - default
status:
  completionTimestamp: "2023-12-13T01:38:52Z"
  expiration: "2024-01-12T01:38:46Z"
  formatVersion: 1.1.0
  phase: Completed
  progress:
    itemsBackedUp: 26
    totalItems: 26
  startTimestamp: "2023-12-13T01:38:46Z"
  version: 1
  warnings: 1
```
- Create Restore
```shell
velero create restore migrate-restore --from-backup migrate-backup  --namespace-mappings cluster-migrate-test-app:migrate-restore --include-resources pods,deployments,pvc,service,pv
```