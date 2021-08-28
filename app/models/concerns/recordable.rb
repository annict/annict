# frozen_string_literal: true

module Recordable
  extend ActiveSupport::Concern

  included do
    has_one :record, as: :recordable, touch: true
  end
end
