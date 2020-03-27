(dotimes (i 100)
  (print
   (cond
    ((= 0 (mod i 15)) "FizzBuzz")
    ((= 0 (mod i 3)) "Fizz")
    ((= 0 (mod i 5)) "Buzz")
    ('t i))))
