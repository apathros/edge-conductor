#!/bin/bash
#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

APISCHEMAS_DIR=.
PATTERNURL='(?:(?:https?|http|ftp|file|oci):\/\/|www\.|ftp\.)(?:\([-A-Z0-9+\&\@\#\/%=~_|$?!:,.]*\)|[-A-Z0-9+\&\@\#\/%=~_|$?!:,.])*(?:\([-A-Z0-9+\&\@\#\/%=~_|$?!:,.]*\)|[A-Z0-9+\&\@\#\/%=~_|$])'
PATTERNPORT='^((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4}))$'
PATTERNNORMALSTRING='^[a-zA-Z_$][a-zA-Z_.\\-$0-9]*$'
PATTERNFILEPATH='^[a-zA-Z\.\\\/][a-zA-Z0-9\-\_\.\\\/]*$'
PATTERNIPV4='(\\b25[0-5]|\\b2[0-4][0-9]|\\b[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}\\b'
PATTERNIPV6='(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))'
PATTERNMAC='^[a-fA-F0-9]{2}(:[a-fA-F0-9]{2}){5}$'
PATTERNIMAGENAMESTRING='^[a-zA-Z0-9][a-zA-Z_.\\-\/:@0-9]*[a-zA-Z0-9]$'



grep @PATTERNURL@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNURL@/\'$PATTERNURL\'/g"
grep @PATTERNPORT@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNPORT@/\'$PATTERNPORT\'/g"
grep @PATTERNNORMALSTRING@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNNORMALSTRING@/\'$PATTERNNORMALSTRING\'/g" 
grep @PATTERNFILEPATH@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNFILEPATH@/\'$PATTERNFILEPATH\'/g" 
grep @PATTERNIPV4@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNIPV4@/\'$PATTERNIPV4\'/g" 
grep @PATTERNIPV6@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNIPV6@/\'$PATTERNIPV6\'/g" 
grep @PATTERNMAC@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNMAC@/\'$PATTERNMAC\'/g"
grep @PATTERNIMAGENAMESTRING@ --include="*.yml" -rl $APISCHEMAS_DIR | xargs -r sed -i "s/@PATTERNIMAGENAMESTRING@/\'$PATTERNIMAGENAMESTRING\'/g"
