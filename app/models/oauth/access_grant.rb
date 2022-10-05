# frozen_string_literal: true

class Oauth::AccessGrant < ApplicationRecord
  include ::Doorkeeper::Orm::ActiveRecord::Mixins::AccessGrant

  include BatchDestroyable
end
