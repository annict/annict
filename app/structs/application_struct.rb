# frozen_string_literal: true

class ApplicationStruct < Dry::Struct
  def decorate
    ActiveDecorator::Decorator.instance.decorate(self)
  end
end
