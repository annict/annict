# frozen_string_literal: true

describe Creators::LikeCreator, type: :model do
  context "Record (WorkRecord) にLikeするとき" do
    let!(:user) { create(:user, :with_setting) }
    let!(:work) { create(:work) }
    let!(:record) { create(:record, user: user, work: work) }
    let!(:work_record) { create(:work_record, user: user, work: work, record: record) }

    context "Likeしていないとき" do
      it "Likeできること" do
        # Creatorを呼んでいないため、このタイミングでは0件のはず
        expect(Like.count).to eq 0

        # Creatorを呼んでLikeする
        result = Creators::LikeCreator.new(user: user, likeable: record).call

        # likes テーブルにレコードが1件作成されるはず
        expect(Like.count).to eq 1

        like = Like.first

        # Creatorを呼んだときに指定した User の値で保存されていることを確認する
        expect(like.user).to eq user
        # Creatorを呼んだときに指定した Record に紐付く WorkRecord の値で保存されていることを確認する
        expect(like.recipient).to eq work_record
        # Creatorの返り値から、作成した likes テーブルのレコードが取得できることを確認する
        expect(result.like).to eq like
      end
    end

    context "Likeしているとき" do
      # 事前にLikeしておく
      let!(:like) { create(:like, user: user, recipient: work_record) }

      it "新しくLikeせず、すでに保存されているLikeが返ること" do
        # 事前にLikeしているため、Creatorを呼んでいなくても1件存在するはず
        expect(Like.count).to eq 1

        like_before = Like.first

        # Creatorを呼んでLikeしようとする
        result = Creators::LikeCreator.new(user: user, likeable: record).call

        # 新たにレコードは作成せず、1件のみであることを確認する
        expect(Like.count).to eq 1

        like_after = Like.first

        # Creatorを呼んだあとも likes テーブルにレコードに変化が無いことを確認する
        expect(like_after).to eq like_before
        # Creatorの返り値から、もともとあった likes テーブルのレコードが取得できることを確認する
        expect(result.like).to eq like_before
      end
    end
  end

  context "Record (EpisodeRecord) にLikeするとき" do
    let!(:user) { create(:user, :with_setting) }
    let!(:work) { create(:work) }
    let!(:episode) { create(:episode, work: work) }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode) }
    let!(:record) { episode_record.record }

    context "Likeしていないとき" do
      it "Likeできること" do
        # Creatorを呼んでいないため、このタイミングでは0件のはず
        expect(Like.count).to eq 0

        creator = Creators::LikeCreator.new(user: user, likeable: record)
        allow(creator).to receive(:send_notification)

        # Creatorを呼んだとき #send_notification が一度呼ばれることを確認する
        expect(creator).to receive(:send_notification).once

        # Creatorを呼んでLikeする
        result = creator.call

        # likes テーブルにレコードが1件作成されるはず
        expect(Like.count).to eq 1

        like = Like.first

        # Creatorを呼んだときに指定した User の値で保存されていることを確認する
        expect(like.user).to eq user
        # Creatorを呼んだときに指定した Record に紐付く EpisodeRecord の値で保存されていることを確認する
        expect(like.recipient).to eq episode_record
        # Creatorの返り値から、作成した likes テーブルのレコードが取得できることを確認する
        expect(result.like).to eq like
      end
    end

    context "Likeしているとき" do
      # 事前にLikeしておく
      let!(:like) { create(:like, user: user, recipient: episode_record) }

      it "新しくLikeせず、すでに保存されているLikeが返ること" do
        # 事前にLikeしているため、Creatorを呼んでいなくても1件存在するはず
        expect(Like.count).to eq 1

        like_before = Like.first

        creator = Creators::LikeCreator.new(user: user, likeable: record)
        allow(creator).to receive(:send_notification)

        # すでにLikeしているため、 #send_notification は一度も呼ばれないはず
        expect(creator).to receive(:send_notification).exactly(0).times

        # Creatorを呼んでLikeしようとする
        result = creator.call

        # 新たにレコードは作成せず、1件のみであることを確認する
        expect(Like.count).to eq 1

        like_after = Like.first

        # Creatorを呼んだあとも likes テーブルにレコードに変化が無いことを確認する
        expect(like_after).to eq like_before
        # Creatorの返り値から、もともとあった likes テーブルのレコードが取得できることを確認する
        expect(result.like).to eq like_before
      end
    end
  end
end
