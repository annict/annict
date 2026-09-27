# typed: false
# frozen_string_literal: true

describe Deprecated::ChannelsQuery, type: :query do
  context "is_vodがnilのとき" do
    it "すべてのチャンネルを返すこと" do
      vod_channel = create(:channel, :with_vod)
      non_vod_channel = create(:channel)

      channels = Deprecated::ChannelsQuery.new(Channel.all, is_vod: nil).call

      expect(channels.pluck(:id)).to contain_exactly(vod_channel.id, non_vod_channel.id)
    end
  end

  context "is_vodがtrueのとき" do
    it "VODのチャンネルだけを返すこと" do
      vod_channel = create(:channel, :with_vod)
      create(:channel)

      channels = Deprecated::ChannelsQuery.new(Channel.all, is_vod: true).call

      expect(channels.pluck(:id)).to contain_exactly(vod_channel.id)
    end
  end

  context "is_vodがfalseのとき" do
    it "VODではないチャンネルだけを返すこと" do
      create(:channel, :with_vod)
      non_vod_channel = create(:channel)

      channels = Deprecated::ChannelsQuery.new(Channel.all, is_vod: false).call

      expect(channels.pluck(:id)).to contain_exactly(non_vod_channel.id)
    end
  end
end
