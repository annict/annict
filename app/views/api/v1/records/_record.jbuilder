# frozen_string_literal: true

json.id record.id if @params.fields_contain?("#{field_prefix}id")
json.comment record.comment if @params.fields_contain?("#{field_prefix}comment")
json.rating record.rating if @params.fields_contain?("#{field_prefix}rating")
json.is_modified record.modify_comment if @params.fields_contain?("#{field_prefix}is_modified")
json.likes_count record.likes_count if @params.fields_contain?("#{field_prefix}likes_count")
json.comments_count record.comments_count if @params.fields_contain?("#{field_prefix}comments_count")
json.created_at record.created_at if @params.fields_contain?("#{field_prefix}created_at")
