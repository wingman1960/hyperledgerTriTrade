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
export class CreateComponent implements OnInit {
  /**
   * Set our default values
   */
  // public localState = { value: '',
  //                       marbleName:'',
  //                       marbleColor:'',
  //                       marbleSize:'',
  //                       marbleOwner: ''
                        

  // };

  public marbleInputState = { 
        marbleName:'',
        marbleColor:'',
        marbleSize:'50',
        marbleOwner: ''
        

};

  public owners = ['Alan','Bob','Cat']

  public settingsTable = {
            actions: false,
            columns: {
              marbleOwner: {
                title: 'marble Owner'
              },
              marbleName: {
                title: 'marble Name'
              },
              marbleColor: {
                title: 'marble Color'
              },
              marbleSize: {
                title: 'marble Size'
              }
            }
          };

  public dataTable = []; // data container for the table

  public assetsStates = {};

  results: string[];
  /**
   * TypeScript public modifiers
   */
  constructor(
    public appState: AppState,
    public title: Title,
    public http: Http
  ) {}

  public ngOnInit() {
    console.log('hello `Home` component');
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



  postRequest2(url, data) {
    data = {
      "fcn" : "initMarble",
      "args" : ["marble3","green","50","bob"]
    }
    console.log("in postRequest2")
    console.log(data)
    var headers = new Headers(), authtoken = localStorage.getItem('authtoken');
    headers.append("Content-Type", 'application/json');

    // if (authtoken) {
    // headers.append("Authorization", 'Token ' + authtoken)
    // }
    headers.append("Accept", 'application/json');

    var requestoptions = new RequestOptions({
              method: RequestMethod.Post,
              url: "http://localhost:3000/invoke",
              headers: headers,
              body: JSON.stringify(data)
    })

    return this.http.request(new Request(requestoptions))
    .map((res: Response) => {
      console.log("in postRequest2-sub")
        if (res) {
            return { status: res.status, json: res.json() }
        }
    });
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


  postNewMarble(){
    var baseUrl = "http://localhost:3000/invoke"
    var body = {
      "fcn" : "initMarble",
      "args" : [this.marbleInputState.marbleName,this.marbleInputState.marbleColor,this.marbleInputState.marbleSize,this.marbleInputState.marbleOwner]
    }
    this.postRequest(baseUrl, body).subscribe((data) => {
      
      console.log(data)
      
      return this.getAssetsStates(this.marbleInputState.marbleOwner)
    });
    
  }


  getAssetsStates(owner){
    var baseUrl = "http://localhost:3000/query"
    var params = {
      "fcn": "queryMarblesByOwner",
      "args": owner
    }
    
    this.getRequest(baseUrl, params).subscribe((data) => {
      this.assetsStates[owner] = data 
      console.log(data)
      return this.updateTable()
    });
    
  }

  getAllAssetsStates(){
    var promises = []
    var owners = this.owners
    for (var i = 0; i < owners.length; i++) {
      console.log(owners[i])
      this.getAssetsStates(owners[i])
      // promises.push (this.getAssetsStates(owners[i]));
    } 
    return Promise.all(promises)
  }

  updateTable(){
    this.dataTable = []; //clear table

    var assetsState = this.assetsStates
    for(var key in assetsState) {
      console.log(key)
      console.log(assetsState[key][0])
      var j = assetsState[key].length
      for (var i = 0; i < j; i++) {
        console.log(assetsState[key][i])
        var temp = assetsState[key][i]
        let record = temp["Record"]
        console.log(assetsState[key][i])
        var tempDict = {
          marbleName: '',
          marbleColor: '',
          marbleSize: '',
          marbleOwner: ''
        }
        console.log(tempDict)
        tempDict.marbleName = record.name
        tempDict.marbleColor = record.color
        tempDict.marbleSize = record.size
        tempDict.marbleOwner = record.owner
        console.log(record)
        this.dataTable.push(tempDict);
    }
    console.log(this.dataTable)
    }
  }
  
}
