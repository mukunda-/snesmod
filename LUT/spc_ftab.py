# hello, world of python!

output = open("ftab.txt", "w")
output.write("LUT_FTAB:")

#for i in range(0,768):
#    if( (i % 16) == 0 ):
#        output.write( "\n        .word " )
#    output.write( str( int(round((1070.464*8) * 2**(i/768.0)) )) )
#    if( (i % 16) != 15 ):
#        output.write( ", " )

for i in range(0,768):
    if( (i % 16) == 0 ):
        output.write( "\n        .word " )
    output.write( hex( int(round((1070.464*8) * 2**(i/768.0)) )).replace("0x","").upper().zfill(5) + "h" )
    if( (i % 16) != 15 ):
        output.write( ", " )

output.write( "\n\n" )

output.close()

