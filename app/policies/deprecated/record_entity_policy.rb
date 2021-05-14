# frozen_string_literal: true

module Deprecated
  class RecordEntityPolicy < ApplicationPolicy
    def update?
      user&.id == record.user.database_id
    end

    def destroy?
      update?
    end
  end
end
