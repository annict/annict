# frozen_string_literal: true

class CreateItemJob < ApplicationJob
  queue_as :default

  def perform(user_id, resource_type, resource_id, asin)
    user = User.find(user_id)
    client = Annict::Amazon::Client.new
    amazon_item = client.items.fetch_by_asin(asin)

    ActiveRecord::Base.transaction do
      item = Item.where(asin: asin).first_or_create! do |i|
        i.title = amazon_item.title
        i.detail_page_url = amazon_item.detail_page_url
        i.ean = amazon_item.ean
        i.amount = amazon_item.amount
        i.currency_code = amazon_item.currency_code
        i.offer_amount = amazon_item.offer_amount
        i.offer_currency_code = amazon_item.offer_currency_code
        i.release_on = if amazon_item.release_date.present?
          Date.parse(amazon_item.release_date)
        end
        i.manufacturer = amazon_item.manufacturer
        i.image = if amazon_item.images.present?
          Down.open(amazon_item.images.first[:url])
        end
      end

      resource = resource_type.constantize.find(resource_id)
      resource.resource_items.where(item: item).first_or_create! do |ri|
        ri.user = user
        case resource
        when Episode
          ri.work = resource.work
        end
      end
    end
  end
end
