#!/bin/sh

alias antlr4='java -Xmx500M -cp "antlr-4.13.1-complete.jar:$CLASSPATH" org.antlr.v4.Tool'
antlr4 -Dlanguage=Go -no-visitor -package grammar ACSL.g4 -o ../generated/grammar/acsl
antlr4 -Dlanguage=Go -no-visitor -package grammar C.g4 -o ../generated/grammar/c
