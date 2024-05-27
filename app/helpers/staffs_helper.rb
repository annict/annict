# typed: false
# frozen_string_literal: true

module StaffsHelper
  def staff_resource_path(staff)
    case staff.resource_type
    when "Person" then person_path(staff.resource_id)
    when "Organization" then organization_path(staff.resource_id)
    end
  end
end
