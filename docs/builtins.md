# Builtins

Evy provides built-in functions and events that allow for user
interaction, graphics, animation, mathematical operations, and more.

Functions are self-contained blocks of code that perform a specific
task. Events are notifications that are sent to a program when
something happens, such as when a user moves the mouse or presses a
key.

## Table of Contents

1. [**Input and Output**](builtins_01_io.md)  
   [print](builtins_01_io.md#print), [read](builtins_01_io.md#read), [cls](builtins_01_io.md#cls), [printf](builtins_01_io.md#printf)
2. [**Types**](builtins_02_types.md)  
   [len](builtins_02_types.md#len), [typeof](builtins_02_types.md#typeof)
3. [**Map**](builtins_03_map.md)  
   [del](builtins_03_map.md#del), [has](builtins_03_map.md#has)
4. [**Program Control**](builtins_04_pcontrol.md)  
   [exit](builtins_04_pcontrol.md#exit), [sleep](builtins_04_pcontrol.md#sleep)
5. [**Conversion**](builtins_05_conv.md)  
   [str2num](builtins_05_conv.md#str2num), [str2bool](builtins_05_conv.md#str2bool)
6. [**Error**](builtins_06_error.md)  
   [Fatal Errors](builtins_06_error.md#fatal-errors), [Non-fatal Errors](builtins_06_error.md#non-fatal-errors)
7. [**String**](builtins_07_string.md)  
   [sprint](builtins_07_string.md#sprint), [sprintf](builtins_07_string.md#sprintf), [join](builtins_07_string.md#join), [split](builtins_07_string.md#split), [upper](builtins_07_string.md#upper), [lower](builtins_07_string.md#lower), [index](builtins_07_string.md#index), [startswith](builtins_07_string.md#startswith), [endswith](builtins_07_string.md#endswith), [trim](builtins_07_string.md#trim), [replace](builtins_07_string.md#replace)
8. [**Random**](builtins_08_random.md)  
   [rand](builtins_06_error.md#rand), [rand1](builtins_06_error.md#rand1)
9. [**Math**](builtins_09_math.md)  
   [min](builtins_09_math.md#min), [max](builtins_09_math.md#max), [floor](builtins_09_math.md#floor), [ceil](builtins_09_math.md#ceil), [round](builtins_09_math.md#round), [pow](builtins_09_math.md#pow), [log](builtins_09_math.md#log), [sqrt](builtins_09_math.md#sqrt), [sin](builtins_09_math.md#sin), [cos](builtins_09_math.md#cos), [atan2](builtins_09_math.md#atan2)
10. [**Graphics**](builtins_10_graphics.md)  
    [move](builtins_10_graphics.md#move), [line](builtins_10_graphics.md#line), [rect](builtins_10_graphics.md#rect), [circle](builtins_10_graphics.md#circle), [color](builtins_10_graphics.md#color), [colour](builtins_10_graphics.md#colour), [width](builtins_10_graphics.md#width), [clear](builtins_10_graphics.md#clear), [grid](builtins_10_graphics.md#grid), [gridn](builtins_10_graphics.md#gridn), [poly](builtins_10_graphics.md#poly), [ellipse](builtins_10_graphics.md#ellipse), [stroke](builtins_10_graphics.md#stroke), [fill](builtins_10_graphics.md#fill), [dash](builtins_10_graphics.md#dash), [linecap](builtins_10_graphics.md#linecap), [text](builtins_10_graphics.md#text), [font](builtins_10_graphics.md#font)
11. [**Event Handlers**](builtins_11_events.md)  
    [key](builtins_11_events.md#key), [down](builtins_11_events.md#down), [up](builtins_11_events.md#up), [move](builtins_11_events.md#move), [animate](builtins_11_events.md#animate), [input](builtins_11_events.md#input)
