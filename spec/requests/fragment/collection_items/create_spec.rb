# typed: false
# frozen_string_literal: true

RSpec.describe "POST /fragment/works/:work_id/collection_items", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    work = FactoryBot.create(:work)

    post "/fragment/works/#{work.id}/collection_items", params: {
      forms_collection_item_form: {
        collection_id: "dummy"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "有効なパラメータでコレクションアイテムが作成されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
    expect(response).to redirect_to(fragment_new_collection_item_path(work))

    collection_item = CollectionItem.last
    expect(collection_item.user).to eq(user)
    expect(collection_item.work).to eq(work)
    expect(collection_item.collection).to eq(collection)
  end

  it "無効なコレクションIDでバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: "invalid"
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("コレクション")
  end

  it "存在しないworkIDで404エラーになること" do
    user = FactoryBot.create(:registered_user)
    collection = FactoryBot.create(:collection, user: user)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/999999/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "他のユーザーのコレクションを指定するとバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    other_collection = FactoryBot.create(:collection, user: other_user)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: other_collection.id
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
  end

  it "既に同じワークがコレクションに追加されている場合、エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)

    # 既存のコレクションアイテムを作成
    CollectionItem.create!(
      user: user,
      work: work,
      collection: collection
    )

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to raise_error(ActiveRecord::RecordNotUnique)
  end
end
