# frozen_string_literal: true

describe CreateEpisodeRecordService, type: :service do
  let(:user) { create :registered_user }

  context "正常系" do
    let(:episode) { create :episode }
    let(:anime) { episode.work }

    it "エピソードへの記録が作成できること" do
      # サービスクラスを呼んでいないので、各レコードは0件のはず
      expect(Record.count).to eq 0
      expect(EpisodeRecord.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      # サービスクラスを呼ぶ
      result = CreateEpisodeRecordService.new(
        user: user,
        episode: episode,
        rating: "good",
        comment: "にぱー",
        share_to_twitter: false
      ).call

      # サービスクラスからエラーは返らないはず
      expect(result.errors.length).to eq 0

      # サービスクラスを呼んだので、各レコードが1件ずつ作成されるはず
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

      expect(episode_record.body).to eq "にぱー"
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
        # 感想付きの記録が直前にある
        let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "はうー") }
        let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
        let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

        it "ActivityGroup が新たに作成されること" do
          expect(Record.count).to eq 1
          expect(EpisodeRecord.count).to eq 1
          expect(ActivityGroup.count).to eq 1
          expect(Activity.count).to eq 1
          expect(user.share_record_to_twitter?).to eq false

          # サービスクラスを呼ぶ
          CreateEpisodeRecordService.new(
            user: user,
            episode: episode,
            rating: "good",
            comment: "にぱー", # 感想付きの記録を新たにする
            share_to_twitter: false
          ).call

          expect(ActivityGroup.count).to eq 2 # ActivityGroup が新たに作成されるはず
          expect(Activity.count).to eq 2

          episode_record = user.episode_records.last
          activity_group = user.activity_groups.last
          activity = user.activities.last

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
        # 感想無しの記録が直前にある
        let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "") }
        let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: false) }
        let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

        it "ActivityGroup が新たに作成されないこと" do
          expect(Record.count).to eq 1
          expect(EpisodeRecord.count).to eq 1
          expect(ActivityGroup.count).to eq 1
          expect(Activity.count).to eq 1
          expect(user.share_record_to_twitter?).to eq false

          # サービスクラスを呼ぶ
          CreateEpisodeRecordService.new(
            user: user,
            episode: episode,
            rating: "good",
            comment: "", # 感想無しの記録を新たにする
            share_to_twitter: false
          ).call

          expect(ActivityGroup.count).to eq 1 # ActivityGroup は新たに作成されないはず
          expect(Activity.count).to eq 2

          episode_record = user.episode_records.last
          activity_group = user.activity_groups.first
          activity = user.activities.last

          expect(activity_group.itemable_type).to eq "EpisodeRecord"
          expect(activity_group.single).to eq false

          expect(activity.itemable).to eq episode_record
          # もともとあった ActivityGroup に紐付くはず
          expect(activity.activity_group_id).to eq activity_group.id
        end
      end
    end
  end

  context "異常系" do
    let(:episode) { create :episode }

    context "バリデーションエラーになったとき" do
      it "エラー内容を返すこと" do
        # エラーになってもならなくても最初は各レコードは0件のはず
        expect(Record.count).to eq 0
        expect(EpisodeRecord.count).to eq 0
        expect(ActivityGroup.count).to eq 0
        expect(Activity.count).to eq 0
        expect(user.share_record_to_twitter?).to eq false

        # サービスクラスを呼ぶ
        result = CreateEpisodeRecordService.new(
          user: user,
          episode: episode,
          rating: "good",
          comment: "a" * (1_048_596 + 1), # 文字数制限 (1,048,596文字) 以上の感想を書く
          share_to_twitter: false
        ).call

        # サービスクラスを呼んでもバリデーションエラーになるので、各レコードは0件のままのはず
        expect(Record.count).to eq 0
        expect(EpisodeRecord.count).to eq 0
        expect(ActivityGroup.count).to eq 0
        expect(Activity.count).to eq 0
        expect(user.share_record_to_twitter?).to eq false

        # サービスクラスからエラー内容が受け取れること
        expect(result.errors.length).to eq 1
        expect(result.errors.first.message).to eq "感想は1048596文字以内で入力してください"
      end
    end
  end
end
