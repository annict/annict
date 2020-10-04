# frozen_string_literal: true

class ApplicationComponent < ViewComponent::Base
  private

  def policy(user_entity, resource_entity)
    @policy ||= Pundit::PolicyFinder.new(resource_entity).policy.new(user_entity, resource_entity)
  end
end
