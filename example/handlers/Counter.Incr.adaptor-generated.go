
package handlers

//code generated by 'go generate', do not edit

import (
	"github.com/jargv/plumbus"
	"net/http"
	"reflect"
	"encoding/json"
	"fmt"
	"log"
)

// avoid unused import errors
var _ json.Delim
var _ log.Logger
var _ fmt.Formatter

func init(){
	var dummy func(
		
	)(
		
			*Counter,
		
	)

	typ := reflect.TypeOf(dummy)
	plumbus.RegisterAdaptor(typ, func(handler interface{}) http.HandlerFunc {
		callback := handler.(func(
			
		)(
			
				*Counter,
			
		))

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){
			
			

			
			
				result0  := 
			

			callback(
				
			)

			
			

			
				
					
						{
							if err := json.NewEncoder(res).Encode(result0); err != nil {
								plumbus.HandleResponseError(res, req, err)
								return
							}
						}
					
				
			
		})
	})
}