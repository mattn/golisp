(defmacro mac (x) (print x))
(mac 'a)
(defmacro negate (x) (cons '- (cons x nil)))
(print (negate 10))

(defmacro m nil)
(print (m))

(defmacro m (x) (print x) (list '+ x 'y))
(setq y 1)
(print (m (+ 10 20)))

(defun x (y) (defmacro m (x) (list '+ x y)))
(x 10)
(print (m 20))

(defmacro m (x) (list '+ x 'x))
(setq x 10)
(print (m 20))
