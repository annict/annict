# frozen_string_literal: true

describe AddReactionService, type: :service do
  context "Record (WorkRecord) にリアクションを付けるとき" do
    let!(:user) { create(:user, :with_setting) }
    let!(:work) { create(:work) }
    let!(:record) { create(:record, user: user, work: work) }
    let!(:work_record) { create(:work_record, user: user, work: work, record: record) }

    context "リアクションを付けていないとき" do
      it "リアクションが付けられること" do
        # AddReactionService#call を呼んでいないため、このタイミングでは0件のはず
        expect(Like.count).to eq 0

        # AddReactionService#call を呼んでリアクションを付ける
        result = AddReactionService.new(user: user, reactable: record).call

        # likes テーブルにレコードが1件作成されるはず
        expect(Like.count).to eq 1

        like = Like.first

        # AddReactionService#call を呼んだときに指定した User の値で保存されていることを確認する
        expect(like.user).to eq user
        # AddReactionService#call を呼んだときに指定した Record に紐付く WorkRecord の値で保存されていることを確認する
        expect(like.recipient).to eq work_record
        # AddReactionService#call の返り値から、作成した likes テーブルのレコードが取得できることを確認する
        expect(result.reaction).to eq like
      end
    end

    context "リアクションを付いているとき" do
      # 事前にリアクションを付けておく
      let!(:like) { create(:like, user: user, recipient: work_record) }

      it "新しくリアクションは付けず、すでに保存されているリアクションが返ること" do
        # 事前にリアクションを付けているため、AddReactionService#call を呼んでいなくても1件存在するはず
        expect(Like.count).to eq 1

        like_before = Like.first

        # AddReactionService#call を呼んでリアクションを付けようとする
        result = AddReactionService.new(user: user, reactable: record).call

        # 新たにレコードは作成せず、1件のみであることを確認する
        expect(Like.count).to eq 1

        like_after = Like.first

        # AddReactionService#call を呼んだあとも likes テーブルにレコードに変化が無いことを確認する
        expect(like_after).to eq like_before
        # AddReactionService#call の返り値から、もともとあった likes テーブルのレコードが取得できることを確認する
        expect(result.reaction).to eq like_before
      end
    end
  end

  context "Record (EpisodeRecord) にリアクションを付けるとき" do
    let!(:user) { create(:user, :with_setting) }
    let!(:work) { create(:work) }
    let!(:episode) { create(:episode, work: work) }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode) }
    let!(:record) { episode_record.record }

    context "リアクションを付けていないとき" do
      it "リアクションが付けられること" do
        # AddReactionService#call を呼んでいないため、このタイミングでは0件のはず
        expect(Like.count).to eq 0

        service = AddReactionService.new(user: user, reactable: record)
        allow(service).to receive(:send_notification)

        # AddReactionService#call を呼んだとき AddReactionService#send_notification が一度呼ばれることを確認する
        expect(service).to receive(:send_notification).once

        # AddReactionService#call を呼んでリアクションを付ける
        result = service.call

        # likes テーブルにレコードが1件作成されるはず
        expect(Like.count).to eq 1

        like = Like.first

        # AddReactionService#call を呼んだときに指定した User の値で保存されていることを確認する
        expect(like.user).to eq user
        # AddReactionService#call を呼んだときに指定した Record に紐付く EpisodeRecord の値で保存されていることを確認する
        expect(like.recipient).to eq episode_record
        # AddReactionService#call の返り値から、作成した likes テーブルのレコードが取得できることを確認する
        expect(result.reaction).to eq like
      end
    end

    context "リアクションを付いているとき" do
      # 事前にリアクションを付けておく
      let!(:like) { create(:like, user: user, recipient: episode_record) }

      it "新しくリアクションは付けず、すでに保存されているリアクションが返ること" do
        # 事前にリアクションを付けているため、AddReactionService#call を呼んでいなくても1件存在するはず
        expect(Like.count).to eq 1

        like_before = Like.first

        service = AddReactionService.new(user: user, reactable: record)
        allow(service).to receive(:send_notification)

        # すでにリアクションが付いているため、AddReactionService#send_notification は一度も呼ばれないはず
        expect(service).to receive(:send_notification).exactly(0).times

        # AddReactionService#call を呼んでリアクションを付けようとする
        result = service.call

        # 新たにレコードは作成せず、1件のみであることを確認する
        expect(Like.count).to eq 1

        like_after = Like.first

        # AddReactionService#call を呼んだあとも likes テーブルにレコードに変化が無いことを確認する
        expect(like_after).to eq like_before
        # AddReactionService#call の返り値から、もともとあった likes テーブルのレコードが取得できることを確認する
        expect(result.reaction).to eq like_before
      end
    end
  end
end
