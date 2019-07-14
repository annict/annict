# frozen_string_literal: true

module StaffStructDecorator
  def local_accurated_name
    return local_name if local_name == resource.local_name
    "#{local_name} (#{resource.local_name})"
  end

  def local_accurated_name_link
    link_to local_accurated_name, send("#{resource_typename.downcase}_path", resource.annict_id)
  end
end
