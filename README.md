# 使用方法
## Receiver
选择端口和文件接收路径(点Browser打开文件浏览器)。左侧可以点击填入局域网ip(非必要，只是为了能让发送端自动获取自己ip)，如果不填写则是所有局域网广播自身ip。
右侧单选框点击Receive Enable开启接收模式。
## Sender
选择端口和文件接收路径(点Browser打开文件浏览器)。左侧填入接收地址ip(接收端如果已经开启则会自动填入)。点Send File发送文件
# 构建项目
windows
~~~shell
fyne package -os windows
~~~
android
~~~shell
fyne package -os android
~~~