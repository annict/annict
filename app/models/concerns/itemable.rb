# frozen_string_literal: true

module Itemable
  extend ActiveSupport::Concern

  included do
    def item_added?(item)
      items.where(asin: item.asin).present?
    end
  end
end
