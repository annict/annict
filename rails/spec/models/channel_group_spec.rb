# typed: false
# frozen_string_literal: true

RSpec.describe ChannelGroup, type: :model do
  describe "associations" do
    it "has_many :channels" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      channel1 = Channel.create!(
        channel_group:,
        name: "テレビ東京",
        sort_number: 1
      )
      channel2 = Channel.create!(
        channel_group:,
        name: "フジテレビ",
        sort_number: 2
      )

      expect(channel_group.channels).to include(channel1, channel2)
      expect(channel_group.channels.count).to eq(2)
    end
  end

  describe "scopes" do
    it ".published スコープが公開されたチャンネルグループのみを返すこと" do
      published_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      unpublished_group = ChannelGroup.create!(
        name: "未公開グループ",
        sort_number: 2,
        unpublished_at: Time.current
      )

      expect(ChannelGroup.published).to include(published_group)
      expect(ChannelGroup.published).not_to include(unpublished_group)
    end

    it ".unpublished スコープが未公開のチャンネルグループのみを返すこと" do
      published_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      unpublished_group = ChannelGroup.create!(
        name: "未公開グループ",
        sort_number: 2,
        unpublished_at: Time.current
      )

      expect(ChannelGroup.unpublished).not_to include(published_group)
      expect(ChannelGroup.unpublished).to include(unpublished_group)
    end

    it ".without_deleted スコープが削除されていないチャンネルグループのみを返すこと" do
      active_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      deleted_group = ChannelGroup.create!(
        name: "削除済みグループ",
        sort_number: 2,
        deleted_at: Time.current
      )

      expect(ChannelGroup.without_deleted).to include(active_group)
      expect(ChannelGroup.without_deleted).not_to include(deleted_group)
    end

    it ".only_kept スコープが削除されておらず公開されているチャンネルグループのみを返すこと" do
      active_published = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      active_unpublished = ChannelGroup.create!(
        name: "未公開グループ",
        sort_number: 2,
        unpublished_at: Time.current
      )
      deleted_published = ChannelGroup.create!(
        name: "削除済みグループ",
        sort_number: 3,
        deleted_at: Time.current
      )
      deleted_unpublished = ChannelGroup.create!(
        name: "削除済み未公開グループ",
        sort_number: 4,
        deleted_at: Time.current,
        unpublished_at: Time.current
      )

      result = ChannelGroup.only_kept
      expect(result).to include(active_published)
      expect(result).not_to include(active_unpublished)
      expect(result).not_to include(deleted_published)
      expect(result).not_to include(deleted_unpublished)
    end

    it ".deleted スコープが削除されたチャンネルグループのみを返すこと" do
      active_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )
      deleted_group = ChannelGroup.create!(
        name: "削除済みグループ",
        sort_number: 2,
        deleted_at: Time.current
      )

      expect(ChannelGroup.deleted).not_to include(active_group)
      expect(ChannelGroup.deleted).to include(deleted_group)
    end
  end

  describe "#publish" do
    it "未公開のチャンネルグループを公開できること" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        unpublished_at: Time.current
      )

      expect(channel_group.published?).to be false
      channel_group.publish
      expect(channel_group.reload.published?).to be true
      expect(channel_group.unpublished_at).to be_nil
    end

    it "既に公開されているチャンネルグループに対しても正常に動作すること" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.published?).to be true
      channel_group.publish
      expect(channel_group.reload.published?).to be true
    end
  end

  describe "#unpublish" do
    it "公開されているチャンネルグループを未公開にできること" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.published?).to be true
      channel_group.unpublish
      expect(channel_group.reload.published?).to be false
      expect(channel_group.unpublished_at).not_to be_nil
    end

    it "既に未公開のチャンネルグループに対しても正常に動作すること" do
      original_time = 1.day.ago
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        unpublished_at: original_time
      )

      expect(channel_group.published?).to be false
      channel_group.unpublish
      expect(channel_group.reload.published?).to be false
      expect(channel_group.unpublished_at).not_to eq(original_time)
    end
  end

  describe "#published?" do
    it "unpublished_atがnilの場合はtrueを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.published?).to be true
    end

    it "unpublished_atが設定されている場合はfalseを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        unpublished_at: Time.current
      )

      expect(channel_group.published?).to be false
    end
  end

  describe "#not_deleted?" do
    it "deleted_atがnilの場合はtrueを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.not_deleted?).to be true
    end

    it "deleted_atが設定されている場合はfalseを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        deleted_at: Time.current
      )

      expect(channel_group.not_deleted?).to be false
    end
  end

  describe "#deleted?" do
    it "deleted_atがnilの場合はfalseを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.deleted?).to be false
    end

    it "deleted_atが設定されている場合はtrueを返すこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        deleted_at: Time.current
      )

      expect(channel_group.deleted?).to be true
    end
  end

  describe "attributes" do
    it "name属性を持つこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group.name).to eq("地上波")
    end

    it "sort_number属性を持つこと" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 999
      )

      expect(channel_group.sort_number).to eq(999)
    end

    it "unpublished_at属性を持つこと" do
      time = Time.current
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        unpublished_at: time
      )

      expect(channel_group.unpublished_at).to be_within(1.second).of(time)
    end

    it "deleted_at属性を持つこと" do
      time = Time.current
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1,
        deleted_at: time
      )

      expect(channel_group.deleted_at).to be_within(1.second).of(time)
    end
  end

  describe "destroy関連のメソッド" do
    it "destroy_in_batchesメソッドが使用できること" do
      channel_group = ChannelGroup.create!(
        name: "地上波",
        sort_number: 1
      )

      expect(channel_group).to respond_to(:destroy_in_batches)
    end
  end

  describe "継承関係" do
    it "ApplicationRecordを継承していること" do
      expect(ChannelGroup.superclass).to eq(ApplicationRecord)
    end

    it "Unpublishableモジュールが含まれていること" do
      expect(ChannelGroup.included_modules).to include(Unpublishable)
    end

    it "SoftDeletableモジュールが含まれていること（Unpublishable経由）" do
      expect(ChannelGroup.included_modules).to include(SoftDeletable)
    end
  end

  describe "複雑なシナリオ" do
    it "複数のチャンネルを持つチャンネルグループを正しく管理できること" do
      group1 = ChannelGroup.create!(name: "地上波", sort_number: 1)
      group2 = ChannelGroup.create!(name: "BS", sort_number: 2)

      channel1 = Channel.create!(channel_group: group1, name: "テレビ東京", sort_number: 1)
      channel2 = Channel.create!(channel_group: group1, name: "フジテレビ", sort_number: 2)
      channel3 = Channel.create!(channel_group: group2, name: "BS11", sort_number: 1)

      expect(group1.channels.count).to eq(2)
      expect(group2.channels.count).to eq(1)
      expect(group1.channels).to include(channel1, channel2)
      expect(group2.channels).to include(channel3)
    end

    it "未公開かつ削除済みのチャンネルグループが正しくフィルタリングされること" do
      normal = ChannelGroup.create!(name: "通常", sort_number: 1)
      unpublished = ChannelGroup.create!(name: "未公開", sort_number: 2, unpublished_at: Time.current)
      deleted = ChannelGroup.create!(name: "削除済み", sort_number: 3, deleted_at: Time.current)
      both = ChannelGroup.create!(
        name: "未公開かつ削除済み",
        sort_number: 4,
        unpublished_at: Time.current,
        deleted_at: Time.current
      )

      # only_kept は削除されておらず、公開されているもののみ
      expect(ChannelGroup.only_kept).to include(normal)
      expect(ChannelGroup.only_kept).not_to include(unpublished, deleted, both)

      # without_deleted は削除されていないもの全て
      expect(ChannelGroup.without_deleted).to include(normal, unpublished)
      expect(ChannelGroup.without_deleted).not_to include(deleted, both)

      # published は公開されているもの全て
      expect(ChannelGroup.published).to include(normal, deleted)
      expect(ChannelGroup.published).not_to include(unpublished, both)
    end
  end
end
