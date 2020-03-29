(setq pwd (getenv "CISP_BASE_DIR"))
(if (= (length pwd) 0) (setq pwd "."))
(load (concatenate 'string pwd "/example/fizzbuzz.lisp"))
