# frozen_string_literal: true

describe Deprecated::SlotsQuery, type: :query do
  context "when the `order` option is not specified" do
    let!(:slot_1) { create :slot, created_at: Time.new(2019, 12, 1), started_at: Time.new(2019, 12, 31) }
    let!(:slot_2) { create :slot, created_at: Time.new(2019, 12, 2), started_at: Time.new(2019, 12, 30) }
    let!(:slot_3) { create :slot, created_at: Time.new(2019, 12, 3), started_at: Time.new(2019, 12, 29) }

    it "returns slots which are sorted to asc by `created_at` field" do
      slots = Deprecated::SlotsQuery.new(
        Slot.all
      ).call

      expect(slots.pluck(:id)).to match([slot_1.id, slot_2.id, slot_3.id])
    end
  end

  context "when the `order` option is specified" do
    let!(:slot_1) { create :slot, created_at: Time.new(2019, 12, 1), started_at: Time.new(2019, 12, 31) }
    let!(:slot_2) { create :slot, created_at: Time.new(2019, 12, 2), started_at: Time.new(2019, 12, 30) }
    let!(:slot_3) { create :slot, created_at: Time.new(2019, 12, 3), started_at: Time.new(2019, 12, 29) }

    it "returns slots which are sorted by specified field" do
      slots = Deprecated::SlotsQuery.new(
        Slot.all,
        order: Deprecated::SlotsQuery::OrderProperty.new(:started_at, :asc)
      ).call

      expect(slots.pluck(:id)).to match([slot_3.id, slot_2.id, slot_1.id])
    end
  end
end
