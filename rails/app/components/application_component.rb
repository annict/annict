# typed: false
# frozen_string_literal: true

class ApplicationComponent < ViewComponent::Base
  private

  def policy(viewer, resource_entity)
    Pundit::PolicyFinder.new(resource_entity).policy.new(viewer, resource_entity)
  end
end
