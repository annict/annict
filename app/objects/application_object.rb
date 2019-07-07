# frozen_string_literal: true

class ApplicationObject < Dry::Struct
  def decorate
    ActiveDecorator::Decorator.instance.decorate(self)
  end
end
