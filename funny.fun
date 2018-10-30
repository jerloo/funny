echoln('// funny.fun')

a = 1
b = 2
c = a
echoln(c)
assert(c == 1)
d = c + b
echoln(d)
assert(d == 3)

minus(a, b) {
  return b - a
}

e = minus(a, b)
echoln(e)
assert(e == 1)

if a > 0 {
  echoln('if a > 0')
}

fib(n) {
  if n < 2 {
    return n
  }
  return fib(n - 1) + fib(n - 2)
}

r = fib(1)
echoln(r)
r = fib(2)
echoln(r)
r = fib(3)
echoln(r)
r = fib(4)
echoln(r)
r = fib(5)
echoln(r)
r = fib(6)
echoln(r)
r = fib(7)
echoln(r)
r = fib(8)
echoln(r)

person = {
  name = 'jeremaihloo'
  age = 10
}
assert(person.name == 'jeremaihloo')
echoln(person.age)

Object() {
  return {
    name = 'jeremaihloo'
    age = 10
    isAdult() {
      this.age = this.age + 5
      echoln('test')
      echoln(this.age)
      return true
    }
  }
}

obj = Object()
assert(obj.name == 'jeremaihloo')
obj.age = 20
assert(obj.age == 20)
assert(obj.isAdult())
echoln(obj.isAdult())
echoln(obj.age)

arrdemo = [1,2,3]
echoln(arrdemo[2])
assert(arrdemo[2]==3)

echoln('hash')
hashTest = 'haha'
echoln(hashTest)
echoln(hash('test'))
echoln(hash(hashTest))

echoln(max(10,20))