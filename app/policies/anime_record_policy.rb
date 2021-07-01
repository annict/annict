# frozen_string_literal: true

class AnimeRecordPolicy < ApplicationPolicy
  def update?
    user == record.user
  end

  def destroy?
    user == record.user
  end
end
