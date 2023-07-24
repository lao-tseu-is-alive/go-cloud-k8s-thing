WITH old as (SELECT idtypething,
                    name,
                    description,
                    datecreated,
                    idcreator,
                    tablename,
                    idmanagerthing,
                    thedefault,
                    maxidentity,
                    isactive,
                    flag,
                    b4internet,
                    iconeurl,
                    infotypeurl,
                    typegeometrie
             FROM goeland.goeland_type_thing
             WHERE isactive = true
             )
INSERT
INTO go_thing.type_thing (id, name, description,  external_id, table_name,geometry_type,
                 managed_by, _created_at, _created_by )
SELECT old.idtypething, old.name, old.description, old.idtypething, old.tablename,old.typegeometrie,
        old.idmanagerthing,old.datecreated,old.idcreator
FROM old;

