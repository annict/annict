# frozen_string_literal: true

json.id episode.id if @params.fields_contain?("#{field_prefix}id")
json.number episode.raw_number if @params.fields_contain?("#{field_prefix}number")
json.number_text episode.number if @params.fields_contain?("#{field_prefix}number_text")
json.sort_number episode.sort_number if @params.fields_contain?("#{field_prefix}sort_number")
json.title episode.title if @params.fields_contain?("#{field_prefix}title")
json.records_count episode.checkins_count if @params.fields_contain?("#{field_prefix}records_count")
