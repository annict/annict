# frozen_string_literal: true

json.id program.id if params.fields_contain?("#{field_prefix}id")
json.started_at program.started_at if params.fields_contain?("#{field_prefix}started_at")
json.is_rebroadcast program.rebroadcast? if params.fields_contain?("#{field_prefix}is_rebroadcast")
