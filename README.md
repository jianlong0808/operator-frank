# 使用姿势

## 文档教程
[notion笔记](https://green-hail-334.notion.site/operator-frank-0f0703a071e14006adad4e76c040f513)

## 前提
- 本地kubectl能连接集群, 推荐mac环境使用docker-desktop启动k8s集群

## 部署

### 本地调试
启动控制器
```shell
make manifests generate
make
make install
make run
```
部署cr
```shell
kubectl apply -f config/samples/apps_v1_frank.yaml
```


### 集群内部署
启动控制器
```shell
make manifests generate
make
make install
make docker-build docker-push IMG=<image-name>:<tag>
make deploy IMG=<image-name>:<tag>
#检查是否启动成功
kubectl get pod -n frank-system
```
部署cr
```shell
kubectl apply -f config/samples/apps_v1_frank.yaml
```

## 验证
```shell
kubectl get frank
kubectl get frank -o wide
kubectl scale podsbook podsbook-sample  --replicas=3
```