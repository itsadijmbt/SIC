START   1000
00 FIRST   LDA     LENGTH,X 1000
03        STA     BUFFER   1003
06        LDX     FIRST
09LOOP    LDA     BUFFER
0C        ADD     ONE
0F        STA     BUFFER
12        JEQ     ENDPROG
15        J       LOOP
18ENDPROG RSUB
1B LENGTH  WORD    6
1E        WORD    100
21        WORD    101
24        WORD    102
27 BUFFER  RESW    1
2A ONE     WORD    1
2D        END     FIRST

