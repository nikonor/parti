#!/bin/bash 

siege -t 10S -b -c 100 -H 'Content-Type: application/json' 'http://localhost:8087/foo POST { "id": 1, "name": "John Doe", "groups": [ { "id": 1, "grp_name": "Группа #1", "grp_age": 1 }, { "id": 2, "grp_name": "Группа #2", "grp_age": 2 } ] }' 