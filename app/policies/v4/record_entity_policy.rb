# frozen_string_literal: true

module V4
  class RecordEntityPolicy < ApplicationPolicy
    def update?
      user&.id == record.user.database_id
    end

    def destroy?
      update?
    end
  end
end
