# frozen_string_literal: true

module StaffDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || resource&.name.presence || id
    path = db_edit_staff_path(self)
    link_to name, path, options
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
end
