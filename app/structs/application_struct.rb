# frozen_string_literal: true

class ApplicationStruct < Dry::Struct
  transform_types do |type|
    type.required(false)
  end

  def decorate
    ActiveDecorator::Decorator.instance.decorate(self)
  end
end
