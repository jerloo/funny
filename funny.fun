// 这是一个完整的测试
// scope
// a = 1
// a = a + 1
// scopeTest(a){
//     echoln(a)
//     a = a + 1
//     echoln(a)
// }
// scopeTest(a)
// echoln(a)

// fnIfReturnTest(n) {
//     echoln(n)
//     if n > 1 {
//         return n
//     }
//     echoln(n + 1)
// }
// fnIfReturnTest(2)
// echo(3)

gtTest(a,b){
    if a>b{
        return a
    }else{
        return b
    }
}
r = gtTest(2,1)
echoln(r)
// fib(n) {
//     echoln(n)
//     if n < 2 {
//       return n
//     } else {
//       return fib(n - 2) + fib(n - 1)
//     }
// }

// echoln(fib(35))

// hello = 'hello'
// world = 'world'

// hw = hello + ' ' + world

// echo(hw)

// lt = 2 > 1
// lte = 2 >= 1

// headers = {
//     'Content-Type' = 'application/json'
// }



