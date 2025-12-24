-- 堂食支付方式汇总表
select
    case when htr.store_guid isnull then '合计' else max(x.name) end 战区,
    null 城市,
    case when htr.store_guid isnull then null else max(x.username) end 区经理,
    htr.store_guid 门店编码,
    case when htr.store_guid isnull then null else max(htr.store_name) end 门店,
    htr.business_day::text 营业日期,

    @_start
    select
        string_agg(
                concat(
                        'sum(
                            case
                                when htr.payment_type_name = ''',x.payment_type_name,''' then htr.amount - htr.refund_amount
                        end
                    ) as ','"',x.payment_type_name,'"'
                ),','
        )
    from (
         select
             payment_type_name
         from "hst_trade_887f2181-eb06-4d77-b914-7c37c884952c_db".hst_transaction_record,
              LATERAL (
                       select
                           public.getlist(
                                   public.getacl(
                                           $privileges_flag,
                                           ($power_list)::text
                                   ),
                                   ($power_list)::text,
                                   array[
                                       [[ {{STORE_GUID}} -- ]] '-1'
                            ,
                            [[ {{GROUP_STORE_GUID}} -- ]] '-1'
                        ]
                    ) list,
                           public.getacl(
                                   $privileges_flag,
                                   ($power_list)::text
                           ) chmod
                  ) acl
         where
             acl.chmod
           and is_delete = 0
           and state = 4
             [[ and business_day between {{START_TIME}}::date and {{END_TIME}}::date ]]
            and case acl.list
                when 'true' then true
                when 'false' then false
                -- when 'false' then true  --debug
                else (store_guid = any(acl.list::text[]))
            end
            [[ and payment_type_name = any({{PAYMENT_TYPE_NAME}}::text[]) ]]
         group by
             payment_type_name
         order by
             payment_type_name
     ) x
    @_end,

    sum(htr.amount - htr.refund_amount) 合计
from "hst_trade_887f2181-eb06-4d77-b914-7c37c884952c_db".hst_transaction_record htr
    left join (
        select
        mos.external_part_org_id,
        u.username,
        ti2.name
        from team.mapping_org_structure mos -- on htr.store_guid = mos.external_part_org_id
        left join team.team_info ti on mos.holder_id = ti.id
        left join team.team_info ti2 on ti.parent_id = ti2.id
        left join team.user u on ti2.responsible_user = u.id or ti.responsible_user = u.id
    ) x on htr.store_guid = x.external_part_org_id,
    LATERAL (
        select
                public.getlist(
                    public.getacl(
                    $privileges_flag,
                    ($power_list)::text
                    ),
                    ($power_list)::text,
                    array[
                        [[ {{STORE_GUID}} -- ]] '-1'
                        ,
                        [[ {{GROUP_STORE_GUID}} -- ]] '-1'
                    ]
                ) list,
                public.getacl(
                    $privileges_flag,
                    ($power_list)::text
            ) chmod
    ) acl
where
    acl.chmod
  and htr.is_delete = 0
  and htr.state = 4
    [[ and htr.business_day between {{START_TIME}}::date and {{END_TIME}}::date ]]
  and case acl.list
    when 'true' then true
    when 'false' then false
-- when 'false' then true  --debug
    else (htr.store_guid = any(acl.list::text[]))
end
[[ and htr.payment_type_name = any({{PAYMENT_TYPE_NAME}}::text[]) ]]
group by
    grouping sets((htr.store_guid,htr.business_day),())
order by
    case
        when htr.store_guid isnull or htr.store_guid = '' or htr.store_guid = '合计' then 3
        else 5
end,
    5,6;