# frozen_string_literal: true

class StaffDecorator < ApplicationDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || resource&.name.presence || id
    path = h.edit_db_staff_path(self)
    h.link_to name, path, options
  end

  def local_name
    return name if I18n.locale == :ja
    return name_en if name_en.present?
    name
  end

  def local_name_with_old
    return local_name if local_name == resource.decorate.local_name
    "#{local_name} (#{resource.decorate.local_name})"
  end

  def local_name_with_old_link
    path = case resource_type
    when "Person" then h.person_path(resource)
    when "Organization" then h.organization_path(resource)
    end

    h.link_to local_name_with_old, path
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :resource_id
        resource&.name
      when :role
        send(:role_text)
      when :sort_number
        send(:sort_number).to_s
      else
        send(field)
      end

      hash
    end
  end

  def role_name
    return local_role_other if role_value == "other"
    role_text
  end

  def local_role_other
    return role_other_en if I18n.locale != :ja && role_other_en.present?
    role_other
  end
end
