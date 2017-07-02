# frozen_string_literal: true

class ItemPolicy < ApplicationPolicy
  def destroy?(resource)
    resource.resource_items.where(user: user, item: record).present?
  end
end
