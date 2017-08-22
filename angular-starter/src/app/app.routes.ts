import { Routes } from '@angular/router';
import { HomeComponent } from './home';
import { CreateComponent } from './create';
import { TradeComponent } from './trade';
import { NoContentComponent } from './no-content';

import { DataResolver } from './app.resolver';

export const ROUTES: Routes = [
  { path: '',      component: HomeComponent },
  { path: 'home',  component: HomeComponent },
  { path: 'create', component: CreateComponent },
  { path: 'trade', component: TradeComponent },
  { path: '**',    component: NoContentComponent },
];
