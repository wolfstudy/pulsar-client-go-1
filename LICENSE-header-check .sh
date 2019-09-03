#!/bin/bash
find . -name '*.go' -exec sh -c "if grep -q '// Licensed to the Apache Software Foundation (ASF) under one' '{}';then echo '{}' 1 yes;else  echo '{}' 1 no;fi
if  grep -q '// or more contributor license agreements.  See the NOTICE file' '{}';then echo '{}' 2 yes;else echo '{}' 2 no;fi
if  grep -q '// distributed with this work for additional information' '{}';then echo '{}' 3 yes;else echo '{}' 3 no;fi
if  grep -q '// regarding copyright ownership.  The ASF licenses this file' '{}';then echo '{}' 4 yes;else echo '{}' 4 no;fi
if  grep -q '// to you under the Apache License, Version 2.0 (the' '{}';then echo '{}' 5 yes;else echo '{}' 5 no;fi
if  grep -q '// \"License\"); you may not use this file except in compliance' '{}';then echo '{}' 6 yes;else echo '{}' 6 no;fi
if  grep -q '// with the License.  You may obtain a copy of the License at' '{}';then echo '{}' 7 yes;else echo '{}' 7 no;fi
if  grep -q '//' '{}';then echo '{}' 8 yes;else echo '{}' 8 no;fi
if  grep -q '//   http://www.apache.org/licenses/LICENSE-2.0' '{}';then echo '{}' 9 yes;else echo '{}' 9 no;fi
if  grep -q '//' '{}';then echo '{}' 10 yes;else echo '{}' 10 no;fi
if  grep -q '// Unless required by applicable law or agreed to in writing,' '{}';then echo '{}' 11 yes;else echo '{}' 11 no;fi
if  grep -q '// software distributed under the License is distributed on an' '{}';then echo '{}' 12 yes;else echo '{}' 12 no;fi
if  grep -q '// \"AS IS\" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY' '{}';then echo '{}' 13 yes;else echo '{}' 13 no;fi
if  grep -q '// KIND, either express or implied.  See the License for the' '{}';then echo '{}' 14 yes;else echo '{}' 14 no;fi
if  grep -q '// specific language governing permissions and limitations' '{}';then echo '{}' 15 yes;else echo '{}' 15 no;fi
if  grep -q '// under the License.' '{}' ;then echo '{}' 16 yes;else echo '{}' 16 no;fi" \;



