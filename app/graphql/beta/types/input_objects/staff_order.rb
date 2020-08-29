# frozen_string_literal: true

module Beta
  module Types
    module InputObjects
      class StaffOrder < Beta::Types::InputObjects::Base
        argument :field, Beta::Types::Enums::StaffOrderField, required: true
        argument :direction, Beta::Types::Enums::OrderDirection, required: true
      end
    end
  end
end
