(dotimes (i 100)                     ; マクロ. (dotimes (カウンタ変数 繰り返し数 dotimesの返り値) E0 E1 ...)
  (let (num)                         ; (let (V0 V1 V2 ...) E0 E1 E2 ...) 局所変数を定義する
    (setq num (+ i 1))               ; num に i + 1 を代入
    (if (= 0 (mod num 3))            ; Fizzを判断
      (format t "Fizz")              ; Fizzを出力
      nil)                           ; else節(なにもしない)
    (if (= 0 (mod num 5))            ; Buzzを判断
      (format t "Buzz")              ; Buzzを出力
      nil)                           ; else節(なにもしない)
    (if (and (not (= 0 (mod num 5))) ; FizzでもBuzzでもないのを判断
             (not (= 0 (mod num 3))));
      (format t "~a" num))           ; 数字を出力
    (format t " " )))                ; すきま
