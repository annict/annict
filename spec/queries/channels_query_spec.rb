# frozen_string_literal: true

describe ChannelsQuery, type: :query do
  context "when `is_vod` options is nil" do
    let!(:vod_1) { create :channel, vod: true }
    let!(:vod_2) { create :channel, vod: true }
    let!(:not_vod_1) { create :channel, vod: false }
    let!(:not_vod_2) { create :channel, vod: false }

    it "returns all channels" do
      channels = ChannelsQuery.new(
        Channel.all,
        is_vod: nil
      ).call

      expect(channels.pluck(:id)).to contain_exactly(vod_1.id, vod_2.id, not_vod_1.id, not_vod_2.id)
    end
  end

  context "when `is_vod` options is true" do
    let!(:vod_1) { create :channel, vod: true }
    let!(:vod_2) { create :channel, vod: true }
    let!(:not_vod_1) { create :channel, vod: false }
    let!(:not_vod_2) { create :channel, vod: false }

    it "returns channels which are VOD" do
      channels = ChannelsQuery.new(
        Channel.all,
        is_vod: true
      ).call

      expect(channels.pluck(:id)).to contain_exactly(vod_1.id, vod_2.id)
    end
  end

  context "when `is_vod` options is false" do
    let!(:vod_1) { create :channel, vod: true }
    let!(:vod_2) { create :channel, vod: true }
    let!(:not_vod_1) { create :channel, vod: false }
    let!(:not_vod_2) { create :channel, vod: false }

    it "returns channels which are not VOD" do
      channels = ChannelsQuery.new(
        Channel.all,
        is_vod: false
      ).call

      expect(channels.pluck(:id)).to contain_exactly(not_vod_1.id, not_vod_2.id)
    end
  end
end
