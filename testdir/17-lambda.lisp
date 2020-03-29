(print ((lambda (a b) (+ a b)) 3 4))
(setq plus (lambda (a b) (+ a b)))
(print (funcall plus 1 2))
(print (apply plus '(2 3)))

(setq a 10)
(print ((lambda (a) a) 20))
(print a)
