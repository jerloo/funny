// 这是一个完整的测试

echo(fib(35))

hello = 'hello'
world = 'world'

hw = hello + ' ' + world

echo(hw)

lt = 2 > 1
lte = 2 >= 1

fib(n) {
    if n < 2 {
      return n
    }
    return fib(n - 2) + fib(n - 1)
}

headers = {
    'Content-Type' = 'application/json'
}



