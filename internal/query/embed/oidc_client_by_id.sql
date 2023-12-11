--deallocate q;
--prepare q(text, text, boolean) as

with client as (
	select c.instance_id,
		c.app_id, c.client_id, c.client_secret, c.redirect_uris, c.response_types, c.grant_types,
		c.application_type, c.auth_method_type, c.post_logout_redirect_uris, c.is_dev_mode,
		c.access_token_type, c.access_token_role_assertion, c.id_token_role_assertion,
		c.id_token_userinfo_assertion, c.clock_skew, c.additional_origins, a.project_id, a.state
	from projections.apps6_oidc_configs c
	join projections.apps6 a on a.id = c.app_id and a.instance_id = c.instance_id
	where c.instance_id = $1
		and c.client_id = $2
),
roles as (
	select p.project_id, json_agg(p.role_key) as project_role_keys
	from projections.project_roles4 p
	join client c on c.project_id = p.project_id
		and p.instance_id = c.instance_id
	group by p.project_id
),
keys as (
	select identifier as client_id, json_object_agg(id, encode(public_key, 'base64')) as public_keys
	from projections.authn_keys2
	where $3 = true -- when argument is false, don't waste time on trying to query for keys.
		and instance_id = $1
		and identifier = $2
		and expiration > current_timestamp
	group by identifier
),
settings as (
	select instance_id, json_build_object('access_token_lifetime', access_token_lifetime, 'id_token_lifetime', id_token_lifetime) as settings
	from projections.oidc_settings2
	where aggregate_id = $1
		and instance_id = $1
)

select row_to_json(r) as client from (
	select c.*, r.project_role_keys, k.public_keys, s.settings
	from client c
	left join roles r on r.project_id = c.project_id
	left join keys k on k.client_id = c.client_id
	left join settings s on s.instance_id = s.instance_id
) r;

--execute q('230690539048009730', '236647088211951618@tests', true);