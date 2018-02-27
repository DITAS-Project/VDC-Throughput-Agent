package de.tub.ditas.util;

import com.google.gson.FieldNamingPolicy;
import com.google.gson.FieldNamingStrategy;

import java.lang.reflect.Field;

public class atFieldNamingStrategy implements FieldNamingStrategy {

    /**
     * Changes every field that is called timestamp in java to @timestamp in the Json
     *
     * @param field
     * @return
     */
    @Override
    public String translateName(Field field) {
        String fieldName =
                FieldNamingPolicy.IDENTITY.translateName(field);
        if (fieldName.equals("timestamp")) fieldName = "@".concat(fieldName);
        if (fieldName.equals("bytes")) fieldName = "traffic.bytes";
        if (fieldName.equals("component")) fieldName = "traffic.component";
        if (fieldName.equals("date")) fieldName = "traffic.date";
        if (fieldName.equals("time")) fieldName = "traffic.time";
        if (fieldName.equals("interval")) fieldName = "traffic.interval";
        return fieldName;

    }
}