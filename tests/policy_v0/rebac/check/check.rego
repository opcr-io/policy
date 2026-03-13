package rebac.check

import future.keywords.if

# default to a closed system (deny by default)
default allowed := false

# resource context is expected in the following form:
# {
#   "relation": "relation or permission name",
#   "object_type": "object type that carries the relation or permission",
#   "object_id": "id of object instance with type of object_type"
#   "subject_type": "[optional] subject type accessing the object. default is 'user'",
# }
#
# To perform ReBAC checks with a subject that is not a user:
# * set 'identity_context.type' to 'IDENTITY_TYPE_MANUAL'.
# * set `identity_context.identity` to the subject ID.
# * set `resource_context.subject_type` to the subject type.
allowed if {
	ds.check({
		"object_type": input.resource.object_type,
		"object_id": input.resource.object_id,
		"relation": input.resource.relation,
		"subject_type": subject_type,
		"subject_id": subject_id,
	})
}

default subject_type := "user"

# When using IDENTITY_TYPE_MANUAL, the subject type comes from the resource context.
subject_type := input.resource.subject_type if {
	input.identity.type == "IDENTITY_TYPE_MANUAL"
	input.resource.subject_type != ""
}

# When using IDENTITY_TYPE_MANUAL, the subject ID comes from the identity context.
subject_id := input.identity.identity if {
	input.identity.type == "IDENTITY_TYPE_MANUAL"
} else := input.user.id
