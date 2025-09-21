# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /collection_items/:collection_item_id", type: :request do
  it "ログインしているとき、有効なパラメータでコレクションアイテムが更新されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection, body: "元の説明文")

    login_as(user, scope: :user)

    patch "/collection_items/#{collection_item.id}", params: {
      forms_collection_item_form: {
        body: "更新された説明文"
      }
    }

    collection_item.reload
    expect(collection_item.body).to eq("更新された説明文")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、空の説明文でコレクションアイテムが更新されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection, body: "元の説明文")

    login_as(user, scope: :user)

    patch "/collection_items/#{collection_item.id}", params: {
      forms_collection_item_form: {
        body: ""
      }
    }

    collection_item.reload
    expect(collection_item.body).to eq("")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、説明文の文字数が上限を超えた場合にバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    login_as(user, scope: :user)

    long_body = "a" * 1_048_597 # バリデーション上限の1,048,596文字を超える

    patch "/collection_items/#{collection_item.id}", params: {
      forms_collection_item_form: {
        body: long_body
      }
    }

    expect(response.status).to eq(422)
  end

  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    patch "/collection_items/#{collection_item.id}", params: {
      forms_collection_item_form: {
        body: "更新された説明文"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのコレクションアイテムを更新しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    other_collection = FactoryBot.create(:collection, user: other_user)
    other_collection_item = FactoryBot.create(:collection_item, user: other_user, work: work, collection: other_collection)

    login_as(user, scope: :user)

    patch "/collection_items/#{other_collection_item.id

    expect(response.status).to eq(404)
  end

  it "存在しないコレクションアイテムを更新しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    patch "/collection_items/999999", params: {
    forms_collection_item_form: {
    body: "更新された説明文"

    expect(response.status).to eq(404)
  end

  it "削除済みのコレクションアイテムを更新しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)
    collection_item.destroy!

    login_as(user, scope: :user)

    patch "/collection_items/#{collection_item.id

    expect(response.status).to eq(404)
  end
end
