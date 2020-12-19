# frozen_string_literal: true

describe Canary::Mutations::CreateEpisodeRecord do
  let(:user) { create :registered_user }
  let(:token) { create(:oauth_access_token) }
  let(:context) { { viewer: user, doorkeeper_token: token, writable: true } }

  context "正常系" do
    let(:episode) { create :episode }
    let(:anime) { episode.work }
    let(:id) { GraphQL::Schema::UniqueWithinType.encode(episode.class.name, episode.id) }
    let(:comment) { "にぱー" }
    let(:query) do
      <<~GRAPHQL
        mutation {
          createEpisodeRecord(input: {
            episodeId: "#{id}",
            comment: "#{comment}",
            rating: GOOD
          }) {
            record {
              databaseId
            }
          }
        }
      GRAPHQL
    end

    it "エピソードへの記録が作成できること" do
      expect(Record.count).to eq 0
      expect(EpisodeRecord.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      result = Canary::AnnictSchema.execute(query, context: context)

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "createEpisodeRecord", "record", "databaseId")).to_not be_nil

      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      episode_record = user.episode_records.first
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.work_id).to eq anime.id

      expect(episode_record.body).to eq comment
      expect(episode_record.locale).to eq "ja"
      expect(episode_record.rating_state).to eq "good"
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.work_id).to eq anime.id

      expect(activity_group.itemable_type).to eq "EpisodeRecord"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq episode_record
    end

    describe "アクティビティの作成" do
      context "直前の記録に感想が書かれていて、その後に新たに感想付きの記録をしたとき" do
        let(:episode) { create :episode, episode_record_bodies_count: 1 }
        let(:anime) { episode.work }
        let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "はうー") }
        let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
        let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }
        let(:id) { GraphQL::Schema::UniqueWithinType.encode(episode.class.name, episode.id) }
        let(:comment) { "にぱー" }
        let(:query) do
          <<~GRAPHQL
            mutation {
              createEpisodeRecord(input: {
                episodeId: "#{id}",
                comment: "#{comment}",
                rating: GOOD
              }) {
                record {
                  databaseId
                }
              }
            }
          GRAPHQL
        end

        it "ActivityGroup が新たに作成されること" do
          expect(Record.count).to eq 1
          expect(EpisodeRecord.count).to eq 1
          expect(ActivityGroup.count).to eq 1
          expect(Activity.count).to eq 1
          expect(user.share_record_to_twitter?).to eq false

          result = Canary::AnnictSchema.execute(query, context: context)

          expect(result["errors"]).to be_nil
          expect(result.dig("data", "createEpisodeRecord", "record", "databaseId")).to_not be_nil

          expect(Record.count).to eq 2
          expect(EpisodeRecord.count).to eq 2
          # 直前が感想付きの記録の場合は ActivityGroup は新たに作成される
          expect(ActivityGroup.count).to eq 2
          expect(Activity.count).to eq 2
          expect(user.share_record_to_twitter?).to eq false

          record = user.records.last
          episode_record = user.episode_records.last
          activity_group = user.activity_groups.last
          activity = user.activities.last

          expect(record.work_id).to eq anime.id

          expect(episode_record.body).to eq comment
          expect(episode_record.locale).to eq "ja"
          expect(episode_record.rating_state).to eq "good"
          expect(episode_record.episode_id).to eq episode.id
          expect(episode_record.record_id).to eq record.id
          expect(episode_record.work_id).to eq anime.id

          expect(activity_group.itemable_type).to eq "EpisodeRecord"
          expect(activity_group.single).to eq true

          expect(activity.activity_group_id).to eq activity_group.id
          expect(activity.itemable).to eq episode_record
        end
      end

      context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
        let(:user) { create :registered_user }
        let(:episode) { create :episode }
        let(:anime) { episode.work }
        let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "") }
        let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: false) }
        let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }
        let(:id) { GraphQL::Schema::UniqueWithinType.encode(episode.class.name, episode.id) }
        let(:query) do
          <<~GRAPHQL
            mutation {
              createEpisodeRecord(input: {
                episodeId: "#{id}",
                rating: GOOD
              }) {
                record {
                  databaseId
                }
              }
            }
          GRAPHQL
        end

        it "ActivityGroup が新たに作成されないこと" do
          expect(Record.count).to eq 1
          expect(EpisodeRecord.count).to eq 1
          expect(ActivityGroup.count).to eq 1
          expect(Activity.count).to eq 1
          expect(user.share_record_to_twitter?).to eq false

          result = Canary::AnnictSchema.execute(query, context: context)

          expect(result["errors"]).to be_nil
          expect(result.dig("data", "createEpisodeRecord", "record", "databaseId")).to_not be_nil

          expect(Record.count).to eq 2
          expect(EpisodeRecord.count).to eq 2
          # 直前の記録に感想が無いため ActivityGroup は作成されない
          expect(ActivityGroup.count).to eq 1
          expect(Activity.count).to eq 2
          expect(user.share_record_to_twitter?).to eq false

          record = user.records.last
          episode_record = user.episode_records.last
          activity_group = user.activity_groups.first
          activity = user.activities.last

          expect(episode_record.body).to be_nil
          expect(episode_record.locale).to eq "other"
          expect(episode_record.rating_state).to eq "good"
          expect(episode_record.episode_id).to eq episode.id
          expect(episode_record.record_id).to eq record.id
          expect(episode_record.work_id).to eq anime.id

          expect(activity_group.itemable_type).to eq "EpisodeRecord"
          expect(activity_group.single).to eq false

          # すでに作成されていた ActivityGroup に記録のアクティビティが紐付く
          expect(activity.activity_group_id).to eq activity_group.id
          expect(activity.itemable).to eq episode_record
        end
      end
    end
  end
end
