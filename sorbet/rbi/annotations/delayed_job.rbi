# typed: true

# DO NOT EDIT MANUALLY
# This file was pulled from a central RBI files repository.
# Please run `bin/tapioca annotations` to update it.

module Delayed::MessageSending
  sig { params(options: T.nilable(T::Hash[Symbol, T.untyped])).returns(T.self_type) }
  def delay(options = nil); end
end

class Object
  include Delayed::MessageSending
end

class Module
  include Delayed::MessageSendingClassMethods
end
