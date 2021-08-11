# frozen_string_literal: true

module StaffDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || resource&.name.presence || id
    path = db_edit_staff_path(self)
    link_to name, path, options
  end

  def local_name_with_old
    return local_name if local_name == resource.local_name
    "#{local_name} (#{resource.local_name})"
  end

  def local_name_with_old_link
    link_to local_name_with_old, resource_path
  end

  def local_name_link
    link_to local_name, resource_path
  end

  def resource_path
    case resource_type
    when "Person" then person_path(resource_id)
    when "Organization" then organization_path(resource_id)
    end
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
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
