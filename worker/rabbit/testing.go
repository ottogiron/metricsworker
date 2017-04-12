package rabbit

var validPayload = []byte(`{
  "username": "kodingbot",  
  "count": 12412414,    
  "metric": "kite_call" 
}
`)

var invalidPayload = []byte(`{
  "username": "kodingbot"  
  "count": 12412414,    
  "metric": "kite_call" 
}
`)
