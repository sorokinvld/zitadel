import { COMMA, ENTER, SPACE } from '@angular/cdk/keycodes';
import { Location } from '@angular/common';
import { Component, Injector, Type } from '@angular/core';
import { AbstractControl, UntypedFormControl, UntypedFormGroup } from '@angular/forms';
import { MatLegacyChipInputEvent as MatChipInputEvent } from '@angular/material/legacy-chips';
import { ActivatedRoute } from '@angular/router';
import { take } from 'rxjs';
import {
  AddGitHubEnterpriseServerProviderRequest as AdminAddGitHubEnterpriseServerProviderRequest,
  GetProviderByIDRequest as AdminGetProviderByIDRequest,
  UpdateGitHubEnterpriseServerProviderRequest as AdminUpdateGitHubEnterpriseServerProviderRequest,
} from 'src/app/proto/generated/zitadel/admin_pb';
import { Options, Provider } from 'src/app/proto/generated/zitadel/idp_pb';
import {
  AddGitHubEnterpriseServerProviderRequest as MgmtAddGitHubEnterpriseServerProviderRequest,
  GetProviderByIDRequest as MgmtGetProviderByIDRequest,
  UpdateGitHubEnterpriseServerProviderRequest as MgmtUpdateGitHubEnterpriseServerProviderRequest,
} from 'src/app/proto/generated/zitadel/management_pb';
import { AdminService } from 'src/app/services/admin.service';
import { Breadcrumb, BreadcrumbService, BreadcrumbType } from 'src/app/services/breadcrumb.service';
import { GrpcAuthService } from 'src/app/services/grpc-auth.service';
import { ManagementService } from 'src/app/services/mgmt.service';
import { ToastService } from 'src/app/services/toast.service';
import { requiredValidator } from '../../form-field/validators/validators';

import { PolicyComponentServiceType } from '../../policies/policy-component-types.enum';

@Component({
  selector: 'cnsl-provider-github-es',
  templateUrl: './provider-github-es.component.html',
})
export class ProviderGithubESComponent {
  public showOptional: boolean = false;
  public options: Options = new Options().setIsCreationAllowed(true).setIsLinkingAllowed(true);

  public id: string | null = '';
  public updateClientSecret: boolean = false;
  public serviceType: PolicyComponentServiceType = PolicyComponentServiceType.MGMT;
  private service!: ManagementService | AdminService;
  public readonly separatorKeysCodes: number[] = [ENTER, COMMA, SPACE];
  public form!: UntypedFormGroup;

  public loading: boolean = false;

  public provider?: Provider.AsObject;

  constructor(
    private authService: GrpcAuthService,
    private route: ActivatedRoute,
    private toast: ToastService,
    private injector: Injector,
    private _location: Location,
    breadcrumbService: BreadcrumbService,
  ) {
    this.form = new UntypedFormGroup({
      name: new UntypedFormControl('', [requiredValidator]),
      clientId: new UntypedFormControl('', [requiredValidator]),
      clientSecret: new UntypedFormControl('', [requiredValidator]),
      authorizationEndpoint: new UntypedFormControl('', [requiredValidator]),
      tokenEndpoint: new UntypedFormControl('', [requiredValidator]),
      userEndpoint: new UntypedFormControl('', [requiredValidator]),
      scopesList: new UntypedFormControl(['openid', 'profile', 'email'], []),
    });

    this.authService
      .isAllowed(
        this.serviceType === PolicyComponentServiceType.ADMIN
          ? ['iam.idp.write']
          : this.serviceType === PolicyComponentServiceType.MGMT
          ? ['org.idp.write']
          : [],
      )
      .pipe(take(1))
      .subscribe((allowed) => {
        if (allowed) {
          this.form.enable();
        } else {
          this.form.disable();
        }
      });

    this.route.data.pipe(take(1)).subscribe((data) => {
      this.serviceType = data['serviceType'];

      switch (this.serviceType) {
        case PolicyComponentServiceType.MGMT:
          this.service = this.injector.get(ManagementService as Type<ManagementService>);

          const bread: Breadcrumb = {
            type: BreadcrumbType.ORG,
            routerLink: ['/org'],
          };

          breadcrumbService.setBreadcrumb([bread]);
          break;
        case PolicyComponentServiceType.ADMIN:
          this.service = this.injector.get(AdminService as Type<AdminService>);

          const iamBread = new Breadcrumb({
            type: BreadcrumbType.ORG,
            name: 'Instance',
            routerLink: ['/instance'],
          });
          breadcrumbService.setBreadcrumb([iamBread]);
          break;
      }

      this.id = this.route.snapshot.paramMap.get('id');
      if (this.id) {
        this.clientSecret?.setValidators([]);
        this.getData(this.id);
      }
    });
  }

  private getData(id: string): void {
    this.loading = true;
    const req =
      this.serviceType === PolicyComponentServiceType.ADMIN
        ? new AdminGetProviderByIDRequest()
        : new MgmtGetProviderByIDRequest();
    req.setId(id);
    this.service
      .getProviderByID(req)
      .then((resp) => {
        this.provider = resp.idp;
        this.loading = false;
        if (this.provider?.config?.githubEs) {
          this.form.patchValue(this.provider.config.githubEs);
          this.name?.setValue(this.provider.name);
        }
      })
      .catch((error) => {
        this.toast.showError(error);
        this.loading = false;
      });
  }

  public submitForm(): void {
    this.provider ? this.updateGenericOAuthProvider() : this.addGenericOAuthProvider();
  }

  public addGenericOAuthProvider(): void {
    const req =
      this.serviceType === PolicyComponentServiceType.MGMT
        ? new MgmtAddGitHubEnterpriseServerProviderRequest()
        : new AdminAddGitHubEnterpriseServerProviderRequest();

    req.setName(this.name?.value);
    req.setAuthorizationEndpoint(this.authorizationEndpoint?.value);
    req.setTokenEndpoint(this.tokenEndpoint?.value);
    req.setUserEndpoint(this.userEndpoint?.value);
    req.setClientId(this.clientId?.value);
    req.setClientSecret(this.clientSecret?.value);
    req.setScopesList(this.scopesList?.value);
    req.setProviderOptions(this.options);

    this.loading = true;
    this.service
      .addGitHubEnterpriseServerProvider(req)
      .then((idp) => {
        setTimeout(() => {
          this.loading = false;
          this.close();
        }, 2000);
      })
      .catch((error) => {
        this.toast.showError(error);
        this.loading = false;
      });
  }

  public updateGenericOAuthProvider(): void {
    if (this.provider) {
      const req =
        this.serviceType === PolicyComponentServiceType.MGMT
          ? new MgmtUpdateGitHubEnterpriseServerProviderRequest()
          : new AdminUpdateGitHubEnterpriseServerProviderRequest();
      req.setId(this.provider.id);
      req.setName(this.name?.value);
      req.setAuthorizationEndpoint(this.authorizationEndpoint?.value);
      req.setTokenEndpoint(this.tokenEndpoint?.value);
      req.setUserEndpoint(this.userEndpoint?.value);
      req.setClientId(this.clientId?.value);
      req.setClientSecret(this.clientSecret?.value);
      req.setScopesList(this.scopesList?.value);
      req.setProviderOptions(this.options);

      this.loading = true;
      this.service
        .updateGitHubEnterpriseServerProvider(req)
        .then((idp) => {
          setTimeout(() => {
            this.loading = false;
            this.close();
          }, 2000);
        })
        .catch((error) => {
          this.toast.showError(error);
          this.loading = false;
        });
    }
  }

  public close(): void {
    this._location.back();
  }

  public addScope(event: MatChipInputEvent): void {
    const input = event.chipInput?.inputElement;
    const value = event.value.trim();

    if (value !== '') {
      if (this.scopesList?.value) {
        this.scopesList.value.push(value);
        if (input) {
          input.value = '';
        }
      }
    }
  }

  public removeScope(uri: string): void {
    if (this.scopesList?.value) {
      const index = this.scopesList.value.indexOf(uri);

      if (index !== undefined && index >= 0) {
        this.scopesList.value.splice(index, 1);
      }
    }
  }

  public get name(): AbstractControl | null {
    return this.form.get('name');
  }

  public get authorizationEndpoint(): AbstractControl | null {
    return this.form.get('authorizationEndpoint');
  }

  public get tokenEndpoint(): AbstractControl | null {
    return this.form.get('tokenEndpoint');
  }

  public get userEndpoint(): AbstractControl | null {
    return this.form.get('userEndpoint');
  }

  public get clientId(): AbstractControl | null {
    return this.form.get('clientId');
  }

  public get clientSecret(): AbstractControl | null {
    return this.form.get('clientSecret');
  }

  public get issuer(): AbstractControl | null {
    return this.form.get('issuer');
  }

  public get scopesList(): AbstractControl | null {
    return this.form.get('scopesList');
  }
}
