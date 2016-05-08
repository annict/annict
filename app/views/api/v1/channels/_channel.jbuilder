# frozen_string_literal: true

json.id channel.id if params.fields_contain?("#{field_prefix}id")
json.name channel.name if params.fields_contain?("#{field_prefix}name")
