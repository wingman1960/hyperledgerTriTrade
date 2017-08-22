import {
  Component,
  OnInit
} from '@angular/core';

import { AppState } from '../app.service';
import { Title } from './title';
import { XLargeDirective } from './x-large';
import { Http, Response, RequestOptions, Request, RequestMethod, Headers} from '@angular/http'
import 'rxjs/add/operator/map';
import { Observable } from 'rxjs';

import { Ng2SmartTableModule } from 'ng2-smart-table';
import {DropdownModule} from "ngx-dropdown";


@Component({
  /**
   * The selector is what angular internally uses
   * for `document.querySelectorAll(selector)` in our index.html
   * where, in this case, selector is the string 'home'.
   */
  selector: 'create',  // <home></home>
  /**
   * We need to tell Angular's Dependency Injection which providers are in our app.
   */
  providers: [
    Title
  ],
  /**
   * Our list of styles in our component. We may add more to compose many styles together.
   */
  styleUrls: [ './create.component.css' ],
  /**
   * Every Angular template is first compiled by the browser before Angular runs it's compiler.
   */
  templateUrl: './create.component.html'
})
export class TradeComponent implements OnInit {
  /**
   * Set our default values
   */

  public openTradeInputState = { 
        tradeOwner:'',
        wantColor:'',
        wantSize:'',
        willingColor: '',
        willingSize: ''
      };

  public settingsTable = {
            actions: false,
            columns: {
              tradeOwner: {
                title: 'Trade Asker'
              },
              wantColor: {
                title: 'Want Color'
              },
              wantSize: {
                title: 'Want Size'
              },
              willingColor: {
                title: 'Willing Color'
              },
              willingSize: {
                title: 'Willing Size'
              }
            }
          };
  public dataTable = [];

  public openTradesStates = [];

  /**
   * TypeScript public modifiers
   */
  constructor(
    public appState: AppState,
    public title: Title,
    public http: Http
  ) {}

  public ngOnInit() {
    console.log('hello `trade` component');
    /**
     * this.title.getData().subscribe(data => this.data = data);
     */
  }
  // `http://localhost:3000/query?fcn=queryMarblesByOwner&args=tom`
  getRequest(baseUrl, params) {
    
    var tempArray = [];
    for(let key in params) {
      tempArray.push(key + "=" + params[key]);
    }
    var outputParams = tempArray.join("&")
    var Url = baseUrl + '?' + outputParams

    return this.http.get(Url)
    .map((res:Response) => res.json());
  }


 // Add a new comment
  postRequest (url, body) {

  let bodyString = JSON.stringify(body); // Stringify payload
  let headers      = new Headers({ 'Content-Type': 'application/json' }); // ... Set content type to JSON
  // let options       = new RequestOptions({ headers: headers }); // Create a request option
  console.log(bodyString)
  return this.http.post(url, body) // ...using post request
                   .map((res:Response) => res.text()) // ...and calling .json() on the response to return data
                  //  .catch((error:any) => Observable.throw(error.json().error || 'Server error')); //...errors if any
}   


  postNewTrade(){
    var baseUrl = "http://localhost:3000/invoke"
    var body = {
      "fcn" : "openTrade",
      "args" : [this.openTradeInputState.tradeOwner,
                this.openTradeInputState.wantColor,
                this.openTradeInputState.wantSize,
                this.openTradeInputState.willingColor,
                this.openTradeInputState.willingSize]
    }
    this.postRequest(baseUrl, body).subscribe((data) => {
      
      console.log(data)
      
      return this.getOpenTradesStates()
    });
    
  }


  getOpenTradesStates(){
    var baseUrl = "http://localhost:3000/query"
    var params = {
      "fcn": "readOpenTrade",
      "args": ''
    }
    
    this.getRequest(baseUrl, params).subscribe((data) => {
      this.openTradesStates = data["open_trades"] 
      console.log(data)
      return this.updateTable()
    });
    
  }

  clearAllOpenTrades(){
    var baseUrl = "http://localhost:3000/invoke"
    var body = {
      "fcn" : "clearOpenTrades",
      "args" : []
    }
    this.postRequest(baseUrl, body).subscribe((data) => {
      
      console.log(data)
      
      return this.getOpenTradesStates()
    });
  }

  matchTrades(){
    var baseUrl = "http://localhost:3000/invoke"
    var body = {
      "fcn" : "matchTrade",
      "args" : []
    }
    this.postRequest(baseUrl, body).subscribe((data) => {
      
      console.log(data)
      
      return this.getOpenTradesStates()
    });
  }

  matchTriTrades(){
    var baseUrl = "http://localhost:3000/invoke"
    var body = {
      "fcn" : "matchTriTrade",
      "args" : []
    }
    this.postRequest(baseUrl, body).subscribe((data) => {
      
      console.log(data)
      
      return this.getOpenTradesStates()
    });
  }

  updateTable(){
    this.dataTable = []; //clear table

    var openTradesStates = this.openTradesStates
    var j = openTradesStates.length
    for (var i = 0; i < j; i++) {
      var temp = openTradesStates[i]
      var tempDict = {
        tradeOwner: "",
        wantColor: "",
        wantSize: "",
        willingColor: "",
        willingSize: ""
      }
      console.log(tempDict)


      tempDict.tradeOwner= temp.user,
      tempDict.wantColor= temp.want.color,
      tempDict.wantSize= temp.want.size,
      tempDict.willingColor= temp.willing.color,
      tempDict.willingSize= temp.willing.size
      this.dataTable.push(tempDict);
    }
    console.log(this.dataTable)
  }
  
  
}
