# golang 定时器
golang timer

```go
var start = time.Now().Unix()

func main() {
	tw := timer.New(A, 1000) //参数：回调的函数 func(...interface{}) 间隔时间（可不填，默认为1000毫秒）
	tw.Add(1000, "test")     //添加一个定时任务 参数：延时时间 回调函数的参数
	tw.Add(2000, 123456789)
	tw.AddRepeat(3, 1000, "repeat test")         //重复定时任务3次
	tw.AddRepeat(-1, 1000, "always repeat test") //当循环次数为-1时，无限循环
	tw.AddHasFunc(2000, A2, "a2 test")           //指定回调函数A2
	tw.AddRepeatHasFunc(2, 2000, A2, "a2 test")  //重复2次， 指定回调函数A2
	select {}
}

func A(data ...interface{}) {
	fmt.Printf("A data=%v \t time=%v \n", data, time.Now().Unix()-start)
}

func A2(data ...interface{}) {
	fmt.Printf("A2 data=%v \t time=%v \n", data, time.Now().Unix()-start)
}
```