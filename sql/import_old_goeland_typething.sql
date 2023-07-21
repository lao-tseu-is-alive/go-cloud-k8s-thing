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
             FROM goeland_type_thing
             WHERE isactive = true
             )
INSERT
INTO type_thing (id, name, description,  external_id, table_name,geometry_type,
                 managed_by, created_at, created_by )
SELECT old.idtypething, old.name, old.description, old.idtypething, old.tablename,old.typegeometrie,
        old.idmanagerthing,old.datecreated,old.idcreator
FROM old;

